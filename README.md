# 🏏 InstantWicket Backend

The core API engine powering **InstantWicket**—a real-time cricket scoring and statistics tracking application. Built with **Go** and **Gin**, and backed by a **PostgreSQL** relational database.

Live App: [instant-wicket-stryker.vercel.app](https://instant-wicket-stryker.vercel.app/)

---

## 🚀 Tech Stack

*   **Language:** Go (1.22+)
*   **Web Framework:** Gin Web Framework (`github.com/gin-gonic/gin`)
*   **Database:** PostgreSQL 18
*   **Containerization:** Docker & Docker Compose
*   **Deployment:** Render
*   **Health Check & Monitoring:** UptimeRobot

---

## ✨ Features

*   **Live Match Scoring:** Real-time ball-by-ball updates, partnerships, and over statistics.
*   **Player & Team Management:** Tracks player profiles, team rosters, and career statistics.
*   **Database Migrations:** Schema-driven versioning using structured `.sql` migration files.
*   **CORS Enabled:** Configured for seamless cross-origin communication with the frontend client.
*   **Warm Server Health Check:** Endpoint (`/api/ping`) optimized to maintain low latency and eliminate cold starts.

---

## 📁 Repository Structure

```text
├── config/             # Environment & configuration loading
├── database/           # DB initialization, connection logic, and SQL migrations
│   ├── dbHelper/       # Raw query helpers for models
│   └── migrations/     # Schema migration files (.sql)
├── handler/            # HTTP request handlers & API endpoints
├── middleware/         # Auth & request handling middlewares
├── models/             # Data models & structs
├── server/             # Router setup & Gin server initialization
├── utils/              # Helper utilities
├── docker-compose.yml  # Local database setup
└── main.go             # Application entrypoint
