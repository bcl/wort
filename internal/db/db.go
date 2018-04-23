package db

import "github.com/boltdb/bolt"

/*Init creates the database file and default buckets */
func Init(databaseFile *string) (*bolt.DB, error) {
	db, err := bolt.Open(*databaseFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("serialNames"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("readings"))
		if err != nil {
			return err
		}

		return nil
	})
	return db, err
}
