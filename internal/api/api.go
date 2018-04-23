package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// sensorType is the type of sensor reading
type sensorType string

// Sensor types
const (
	TemperatureSensor sensorType = "T"
	HumiditySensor               = "H"
	CounterSensor                = "C"
)

// sensorReading stores the reading and the sensor's serial number
type sensorReading struct {
	Serial      string
	Type        sensorType
	Temperature float32
	Humidity    int `json:"humidity,omitempty"`
	Count       int `json:"count,omitempty"`
}

/* sensorsHandler returns the human readable names for the sensor serial numbers */
func sensorsHandler(db *bolt.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		/* Read the serialNames bucket from the Bolt database and return it as JSON */
		serialNames := make(map[string]string)
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("serialNames"))
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				serialNames[string(k[:])] = string(v[:])
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusOK, serialNames)
		}
	}
	return gin.HandlerFunc(fn)
}

/* readingsHandler returns a range of readins based on the start and end timestamps (Unix format) */
func readingsHandler(db *bolt.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var errorMsgs []string
		start, startErr := strconv.ParseInt(c.Param("start"), 10, 64)
		if startErr != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("ERROR start timestamp: %s", startErr.Error()))
		}

		end, endErr := strconv.ParseInt(c.Param("end"), 10, 64)
		if endErr != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("ERROR end timestamp: %s", endErr.Error()))
		}

		_, limitErr := strconv.ParseInt(c.DefaultQuery("limit", "0"), 10, 64)
		if limitErr != nil {
			errorMsgs = append(errorMsgs, fmt.Sprintf("ERROR limit value: %s", limitErr.Error()))
		}

		sensors := strings.Split(c.Query("sensors"), ",")
		/* Check sensors to make sure they are in the list */
		if len(sensors) > 1 || (len(sensors) == 1 && sensors[0] != "") {
			err := db.View(func(tx *bolt.Tx) error {
				for s := range sensors {
					b := tx.Bucket([]byte("serialNames"))
					v := b.Get([]byte(sensors[s]))
					if v == nil {
						errorMsgs = append(errorMsgs, fmt.Sprintf("ERROR missing sensor: %s", sensors[s]))
					}
				}
				return nil
			})
			if err != nil {
				errorMsgs = append(errorMsgs, err.Error())
			}
		}

		/* Failing any of these results in a BadRequest error with the reason(s) why */
		if len(errorMsgs) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"errors": errorMsgs})
			return
		}

		readings := make(map[string][]*sensorReading)
		err := db.View(func(tx *bolt.Tx) error {
			c := tx.Bucket([]byte("readings")).Cursor()

			start3339 := time.Unix(start, 0).Format(time.RFC3339)
			end3339 := time.Unix(end, 0).Format(time.RFC3339)

			log.Printf("start = %s end = %s", start3339, end3339)
			// Iterate over the 90's.
			for k, v := c.Seek([]byte(start3339)); k != nil && bytes.Compare(k, []byte(end3339)) <= 0; k, v = c.Next() {
				var values []*sensorReading
				if err := json.Unmarshal(v, &values); err != nil {
					log.Printf("Problem decoding %s: %s", v, err)
					continue
				}

				readings[string(k)] = values
			}

			return nil
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
		}

		c.JSON(http.StatusOK, readings)
	}
	return gin.HandlerFunc(fn)
}

/* newReadingsHandler - store new temperature readings into the database */
func newReadingsHandler(db *bolt.DB) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		var readings []*sensorReading
		if err := c.ShouldBindWith(&readings, binding.JSON); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		now := time.Now().Format(time.RFC3339)
		/* Add the reading to Bolt.
		   It uses 3 buckets:
			* serialNames - to map serial numbers to human readable names
			* <serial#> - bucket for each serial number to store readings by timestamp
			* readings - bucket to store all sensor readings by timestamp
		*/
		err := db.Update(func(tx *bolt.Tx) error {
			/* Update the per-sensor buckets and the serialNames bucket */
			for r := range readings {
				log.Printf("%s %0.2f", readings[r].Serial, readings[r].Temperature)

				/* Bucket to hold serial# -> human name mapping */
				snBkt := tx.Bucket([]byte("serialNames"))
				/* TODO - Get the serial# descriptions from someplace */
				if snErr := snBkt.Put([]byte(readings[r].Serial), []byte("Unknown")); snErr != nil {
					return snErr
				}

				/* Bucket for each serial#, key is the timestamp */
				sBkt, err := tx.CreateBucketIfNotExists([]byte(readings[r].Serial))
				if err != nil {
					return err
				}
				encoded, err := json.Marshal(readings[r])
				if err != nil {
					return err
				}
				sErr := sBkt.Put([]byte(now), encoded)
				if sErr != nil {
					return sErr
				}
			}

			/* Store this block of readings indexed by timestamp */
			encoded, err := json.Marshal(readings)
			if err != nil {
				return err
			}
			/* Bucket to hold all the readings, key is the timestamp */
			rdBkt := tx.Bucket([]byte("readings"))
			if rdErr := rdBkt.Put([]byte(now), encoded); rdErr != nil {
				return rdErr
			}

			return nil
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.Status(http.StatusOK)
		}
	}
	return gin.HandlerFunc(fn)
}

func formatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d%02d/%02d", year, month, day)
}

/*Server handles requests for the temperature readings API */
func Server(db *bolt.DB, listenIP *string, listenPort int) {
	router := gin.Default()
	router.POST("/api/new", newReadingsHandler(db))
	router.GET("/api/sensors", sensorsHandler(db))
	router.GET("/api/readings/:start/:end", readingsHandler(db))
	router.Static("/static", "./html/static")

	router.SetFuncMap(template.FuncMap{
		"formatAsDate": formatAsDate,
	})
	router.LoadHTMLFiles("./html/index.tmpl")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", map[string]interface{}{
			"now": time.Date(2017, 07, 01, 0, 0, 0, 0, time.UTC),
		})
	})

	router.Run(fmt.Sprintf("%s:%d", *listenIP, listenPort))
}
