# NAM (Node Application Manager) - Project Context Guide

## Quick Project Overview
**NAM** is a Node Application Manager written in **Go** with a modern web UI that monitors and manages applications across multiple servers. It provides health checking, application deployment tracking, and administrative controls through a clean web interface.

**Architecture Pattern**: 3-Layer Architecture (Data → Service → Handler)  
**Tech Stack**: Go + Gin + HTMX + TailwindCSS + PostgreSQL  
**UI Philosophy**: Server-side rendered with HTMX for dynamic updates  

---

## Architecture Overview

### 1. **Layered Architecture Structure**
```
src/
├── layers/
│   ├── data/           # Data Access Layer (DAO pattern)
│   ├── handler/        # HTTP Handler Layer (Controllers)
│   │   ├── api/rest/v1/   # REST API endpoints
│   │   └── htmx/          # HTMX-specific handlers
│   └── service/        # Business Logic Layer
├── web/
│   ├── static/         # Static assets (CSS, JS, icons)
│   └── templates/      # HTML templates
├── main.go             # Application entry point
├── webserver.go        # Web server initialization & routing
└── configuration.go    # Configuration management
```

### 2. **Core Application Structure**
- **Entry Point**: `main.go` - Initializes database, services, and web server
- **Web Server**: `webserver.go` - Contains all routing logic and middleware setup
- **Configuration**: YAML-based configuration with environment-specific settings
- **Database**: PostgreSQL with migrations using golang-migrate

### 3. **Layer Responsibilities**
- **Data Layer**: Database queries, model definitions, migrations
- **Service Layer**: Business logic, external integrations, background services
- **Handler Layer**: HTTP request handling, response formatting, validation

---

## Web Development Patterns

### 1. **Template System**
- **Engine**: Go's `html/template` with Gin integration
- **Structure**: Modular templates with inheritance
- **Key Templates**:
  - `common/head_includes.html` - Global CSS/JS includes
  - `common/navbar.html` - Navigation with active state management
  - `pages/*.html` - Full page templates
  - `components/*.html` - Reusable components

### 2. **CSS & Styling Conventions**
- **Framework**: TailwindCSS (loaded via CDN in head_includes)
- **Additional Styles**: Minimal custom CSS in `/static/styles.css`
- **Color Scheme**: Professional blue/indigo theme
  - Primary: `indigo-600` for buttons, active states
  - Success: `green-` variants for healthy states
  - Error: `red-` variants for issues
  - Warning: `yellow-` variants for maintenance
  - Neutral: `gray-` variants for inactive/background

### 3. **Component Patterns**
- **Cards**: `bg-white shadow rounded-lg` for content containers
- **Buttons**: `bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md`
- **Navigation**: Consistent active/inactive states with border indicators
- **Forms**: Standard form styling with validation states
- **Status Indicators**: Color-coded badges for health status

### 4. **HTMX Integration**
- **Purpose**: Dynamic updates without full page reloads
- **Patterns**:
  - `hx-get` for loading components dynamically
  - `hx-post` for form submissions
  - `hx-target` with `hx-swap="innerHTML"` for content replacement
  - `hx-trigger` for event-based updates
- **Key HTMX Routes**: `/htmx/*` endpoints in the handler layer
- **Error Handling**: `HX-Redirect` headers for authentication redirects

### 5. **HTMX Extensions & Error Handling**
- **Extensions Available**:
  - `submitjson` - Converts form data to JSON for API endpoints
  - Located in `/static/htmx-extensions.js`
- **JSON Form Submission Pattern**:
  ```html
  <form hx-post="/api/rest/v1/endpoint" 
        hx-ext="submitjson"
        hx-on::after-request="handleResponse(event)">
  ```
- **Global Error Notification System**:
  - **Template**: `template/components/error-notification`
  - **Usage**: Include `{{ template "template/components/error-notification" . }}` in pages
  - **Auto-handling**: Automatically shows errors for failed HTMX requests (400+ status codes)
  - **Auto-hiding**: Automatically hides errors on successful requests
  - **Manual Control**: `showErrorMessage(errorData)` and `hideErrorNotification()` functions available
  - **Features**: JSON parsing, technical details toggle, manual dismissal
- **Simplified Response Handling**:
  ```javascript
  function handleResponse(event) {
      if (event.detail.successful) {
          // Handle success only - errors are automatic
      }
      // No need to handle errors manually
  }
  ```

---

## Database Schema & Models

### 1. **Core Entities**
- **Server**: Physical/virtual servers hosting applications
- **ApplicationDefinition**: Template/blueprint for applications
- **ApplicationInstance**: Specific deployments of applications on servers
- **Healthcheck**: Health check templates and configurations
- **HealthcheckResult**: Results of health check executions
- **User/Role**: Authentication and authorization
- **Action/ActionTemplate**: Deployment and management actions
- **Secret**: Encrypted storage for API keys, passwords, certificates, and tokens

### 2. **Key Relationships**
- Server 1:N ApplicationInstance
- ApplicationDefinition 1:N ApplicationInstance
- ApplicationDefinition 1:1 Healthcheck (optional)
- ApplicationInstance 1:N HealthcheckResult
- User N:1 Role

### 3. **Migration System**
- **Tool**: golang-migrate with PostgreSQL
- **Pattern**: Numbered migrations (0001_base.up.sql, 0001_base.down.sql)
- **Management**: Built-in CLI commands in main.go (`-db versions|drop|newschema`)

---

## API & Routing Patterns

### 1. **Routing Structure**
```go
// Authentication
/login (GET/POST) - Login page and authentication

// Main Pages (requires auth + admin role)
/ or /dashboard - Main dashboard
/servers - Server management
/applications - Application management
/healthchecks - Health check templates
/actions - Action management
/secrets - Secrets management  
/settings - System settings

// API (REST)
/api/rest/v1/* - RESTful API endpoints
├── /servers - Server CRUD
├── /applications - Application CRUD
├── /healthchecks - Health check CRUD
├── /secrets - Encrypted secrets management
└── /actions - Action management

// HTMX Components
/htmx/* - Dynamic component endpoints
├── /applications - Application components
├── /health - Health status components
└── /navbar_user - User info component
```

### 2. **Handler Patterns**
- **Page Handlers**: Return full HTML pages (`ctx.HTML(200, "pages/template", data)`)
- **API Handlers**: Return JSON (`ctx.JSON(200, data)` or error responses)
- **HTMX Handlers**: Return HTML fragments (`ctx.HTML(200, "components/template", data)`)
- **Error Handling**: Consistent error structure with traces in debug mode

### 3. **Authentication & Authorization**
- **Middleware**: `AuthMiddleware()` for authentication checks
- **Authorization**: `RequireRole(pool, "admin")` middleware for role-based access
- **JWT**: Token-based authentication with configurable keys
- **Redirects**: `HX-Redirect` header for HTMX, standard redirects for regular requests

### 4. **Response Patterns**
- **Success**: Appropriate HTTP status with data
- **Client Errors**: 400-level status with error message
- **Server Errors**: 500-level status with error and trace (in debug)
- **HTMX Redirects**: Special `HX-Redirect` header handling

---

## Development Conventions

### 1. **Code Organization**
- **Handlers**: One handler per major entity (Server, Application, etc.)
- **Models**: Defined in `layers/data/models.go` with DAO structs
- **DTOs**: Data Transfer Objects for API request/response
- **Database**: Separate files for each entity's database operations

### 2. **Naming Conventions**
- **Handlers**: `NewEntityHandler`, `GetPageEntity`, `GetEntityById`
- **Database Functions**: `GetEntityById`, `EntityDbInsert`, `EntityDbUpdate`
- **Templates**: `pages/entity.html`, `components/entity-component.html`
- **Routes**: RESTful paths with consistent patterns

### 3. **Error Handling**
- **Database Errors**: Always check and return detailed errors
- **HTTP Errors**: Use helper functions from `handler.go`
- **Validation**: Gin's binding validation with proper error responses
- **Logging**: Structured JSON logging with slog

### 4. **Configuration**
- **Files**: `config.yaml` for main config, `config.example.yaml` for template
- **Environment**: Support for different modes (debug, release, test)
- **Secrets**: Separate secret management for sensitive data

---

## Common Development Tasks

### 1. **Adding a New Page**
1. Create template in `web/templates/pages/`
2. Add handler in appropriate `layers/handler/` file
3. Add route in `webserver.go`
4. Ensure authentication/authorization middleware
5. Add navigation link to `common/navbar.html`

### 2. **Adding New API Endpoint**
1. Add handler in `layers/handler/api/rest/v1/`
2. Define DTOs in `layers/data/models.go`
3. Add database functions in `layers/data/`
4. Add routes in `webserver.go` under REST section
5. Test with appropriate HTTP client

### 3. **Adding HTMX Component**
1. Create component template in `web/templates/components/`
2. Add handler in `layers/handler/htmx/`
3. Add route under `/htmx/` in `webserver.go`
4. Use in pages with appropriate `hx-*` attributes

### 4. **Database Changes**
1. Create migration files in `layers/data/migrations/`
2. Update model structs in `models.go`
3. Add/update database functions in respective files
4. Test migration up and down

---

## Key Files Reference

### Essential Files for Understanding:
- `main.go` - Application bootstrap and initialization
- `webserver.go` - Complete routing and middleware setup
- `layers/data/models.go` - All data structures and DTOs
- `web/templates/common/navbar.html` - Navigation and styling patterns
- `web/templates/pages/dashboard.html` - Modern UI patterns and HTMX usage
- `layers/handler/dashboardHandler.go` - Handler patterns and filtering

### UI & Frontend Files:
- `web/templates/common/error.html` - Error notification component
- `web/static/htmx-extensions.js` - HTMX extensions (submitjson, etc.)
- `web/templates/pages/secrets.html` - Example of modern page with error handling

### Configuration Files:
- `config.yaml` - Main configuration
- `go.mod` - Dependencies and module definition
- `layers/data/migrations/0001_base.up.sql` - Database schema

---

## Quick Reference Commands

```bash
# Build and run
go run . -config config.yaml

# Database migration management
go run . -config config.yaml -db versions  # Show migrations
go run . -config config.yaml -db drop     # Drop all tables
go run . -config config.yaml -db newschema # Apply all migrations

# Build static binary
./build.sh
```

---

## IMPORTANT: When Working with This Codebase

1. **Always follow the 3-layer architecture** - Don't bypass layers
2. **Use HTMX for dynamic updates** - Avoid JavaScript unless necessary
3. **Follow TailwindCSS conventions** - Use utility classes consistently
4. **Maintain template modularity** - Reuse common components
5. **Handle errors automatically** - Include error notification template for automatic error handling
6. **Use HTMX extensions properly** - `hx-ext="submitjson"` for API endpoints expecting JSON
7. **Include error notifications** - Add `{{ template "template/components/error-notification" . }}` to pages for automatic error handling
8. **Test authentication flows** - Ensure proper middleware usage
9. **Consider mobile responsiveness** - Use Tailwind's responsive classes
10. **Follow RESTful API patterns** - Consistent HTTP methods and status codes
11. **Simplified response handling** - Focus on success logic only, errors are handled automatically

### Error Handling Best Practices:
- Always include the error notification template in pages for automatic error handling
- Errors are automatically shown for failed HTMX requests (400+ status codes)
- Errors are automatically hidden on successful requests
- Only handle success logic in custom response handlers
- Manual error control available via `showErrorMessage()` and `hideErrorNotification()` if needed

This document should be referenced at the beginning of each session to understand the project's architecture, conventions, and patterns.