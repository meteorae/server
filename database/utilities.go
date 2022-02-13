package database

import "github.com/rs/zerolog/log"

// Retrives the SQLite version information from the current database.
func GetSQLiteVersion() interface{} {
	var sqliteVersion interface{}

	rows, rowsErr := db.Raw("SELECT sqlite_version()").Rows()
	if rowsErr != nil {
		log.Error().Err(rowsErr)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err())
	}
	defer rows.Close()
	rows.Next()

	err := rows.Scan(&sqliteVersion)
	if err != nil {
		log.Error().Err(err)
	}

	return sqliteVersion
}

// Retrieves the SQLite build information from the current database.
func GetSQLiteBuildInformation() []string {
	var loadedSqliteExtensions []string

	rows, rowsErr := db.Raw("PRAGMA compile_options").Rows()
	if rowsErr != nil {
		log.Error().Err(rowsErr)
	}

	if rows.Err() != nil {
		log.Error().Err(rows.Err())
	}
	defer rows.Close()

	for rows.Next() {
		var extensionRow interface{}

		err := rows.Scan(&extensionRow)
		if err != nil {
			log.Error().Err(err)
		}

		extensionRowString, ok := extensionRow.(string)
		if !ok {
			log.Error().Msg("Could not convert extension row to string")

			continue
		}

		loadedSqliteExtensions = append(loadedSqliteExtensions, extensionRowString)
	}

	return loadedSqliteExtensions
}
