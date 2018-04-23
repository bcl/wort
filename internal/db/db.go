/*Package db contains the database initialization code

  wort - http API server for temperature sensor readings
  Copyright (C) 2018 Brian C. Lane <bcl@brianlane.com>

  This program is free software; you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation; either version 2 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License along
  with this program; if not, write to the Free Software Foundation, Inc.,
  51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
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
