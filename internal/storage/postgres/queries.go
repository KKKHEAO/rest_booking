package postgres

const (
	createBooking = `
		INSERT INTO bookings (id, table_id, guest_name, guests, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	getBooking = `
		SELECT id, table_id, guest_name, guests, starts_at, ends_at, status
		FROM bookings
		WHERE id = $1`
	getListBookings = `
		SELECT id, table_id, guest_name, guests, starts_at, ends_at, status
		FROM bookings
		ORDER BY starts_at`
	updateBookingStatus = `UPDATE bookings SET status = $2 WHERE id = $1`
	createTable         = `INSERT INTO tables (id, number, seats) VALUES ($1, $2, $3)`
	getTable            = `SELECT id, number, seats FROM tables WHERE id = $1`
	getListTables       = `SELECT id, number, seats FROM tables ORDER BY number`
	lockTable           = `SELECT id FROM tables WHERE id = $1 FOR UPDATE`
)
