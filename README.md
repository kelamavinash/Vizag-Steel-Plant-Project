# Vizag-Steel-Plant-Project
# Vizag Steel Delay System

A web-based application designed to log, manage, and analyze equipment delays for various shops (Blast Furnace, Steel Melting Shop, Wire Rod Mill, etc.) at Vizag Steel. This system allows users to track delay durations, allocate work to different agencies, and generate reports.

## Features

- **User Authentication & Role-Based Access Control:** Secure login with session management. Differentiates between regular users and system administrators.
- **Delay Logging & Management:** 
  - Log new equipment delays with details such as shop, equipment, agency, and expected duration.
  - Update delay statuses (`Pending`, `In Progress`, `Work Done`).
- **Master Data Management:** Manage equipment and sub-equipment lists.
- **Reporting & Analytics:** 
  - View delay reports with filtering options.
  - Export delay data to CSV.
  - Visualize delay data using charts.
- **Bulk Data Import:** Import legacy delay records from CSV files (`sample_delays_data.csv`).
- **Responsive UI:** The frontend uses Go HTML templates with modern styling (Tailwind CSS classes).

## Tech Stack

- **Backend:** [Go](https://golang.org/) (Golang)
- **Web Framework:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database:** [SQLite](https://sqlite.org/index.html) (via `github.com/glebarez/sqlite`)
- **ORM:** [GORM](https://gorm.io/)
- **Session Management:** `gin-contrib/sessions` (Cookie-based)
- **Frontend:** Go HTML Templates (`html/template`)

## Prerequisites

- [Go](https://go.dev/doc/install) (Version 1.26 or higher recommended)

## Installation & Setup

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd vizag-steel-delay-system
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

4. **Access the application:**
   Open your browser and navigate to `http://localhost:8090`.

   **Default Admin Credentials (Seeded):**
   - **Emp No:** `1111`
   - **Password:** `test`

   *Note: On the first run, the application will automatically create `project.db`, perform database migrations, and seed initial data (Master Admin, Equipment Data, and sample delays from `sample_delays_data.csv` if available).*

## Project Structure

```
├── handlers/              # HTTP request handlers and business logic
│   ├── auth.go            # Login, logout, and session middleware
│   ├── entry.go           # Delay entry creation and status updates
│   ├── importer.go        # CSV import logic
│   └── reports.go         # Reporting, charting, and CSV export logic
├── models/                # Database models (User, EqptMaster, DelayData)
├── templates/             # HTML templates (base, login, entry, reports, users)
├── main.go                # Application entry point and router setup
├── go.mod / go.sum        # Go module dependencies
└── sample_delays_data.csv # Sample data for initial database seeding
```

## Database Models

- **User:** Manages employee login credentials, roles (`sys_admin`, etc.), and active status.
- **EqptMaster:** Stores shop codes, equipment, and sub-equipment hierarchies.
- **DelayData:** Core model storing all delay records including timestamps, allocated durations, actual durations, statuses, and assigned agencies (operations, electrical, mechanical, shutdown).

## API & Routes Overview

- **Auth:** `GET /login`, `POST /login`, `GET /logout`
- **Delay Management:** `GET /entry`, `POST /entry`, `POST /api/start-work/:id`, `POST /api/mark-done/:id`
- **Reports:** `GET /reports`, `GET /api/reports/data`, `GET /api/reports/chart`, `GET /api/reports/export`
- **Data Import:** `POST /api/import-csv`
- **Admin (sys_admin):** `GET /users`, `POST /users/add`, `POST /users/role`, `POST /users/toggle`

## License

[MIT License](LICENSE) 
