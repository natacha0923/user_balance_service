package migrations

import "database/sql"

// I don't know a good way to do this with docker-compose 🤷‍
const query = `
CREATE TABLE IF NOT EXISTS user_balance (
  user_id     INT8        NOT NULL,
  balance     INT8        NOT NULL,
  PRIMARY KEY (user_id)
);

INSERT INTO user_balance VALUES (1,50) ON CONFLICT DO NOTHING;
INSERT INTO user_balance VALUES (2,100) ON CONFLICT DO NOTHING;
`

func Run(db *sql.DB) error {
	_, err := db.Exec(query)
	return err
}
