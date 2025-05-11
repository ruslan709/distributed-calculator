## Frontend part for “Distributed Calculator”
## Description
This is a web interface for a distributed calculator with support for registration, authorization, sending expressions to calculate. The frontend is written in pure HTML, CSS and JavaScript, no assembly required.
## Features
* User registration and login (JWT)
* Sending arithmetic expressions for calculation
* Clearing calculation history
* Viewing the status of servers (orchestrator and calculators)
* Automatic updating of results and server statuses
## Project structure
* frontend/
* ├─── index.html ## Application home page
* ├──── styles.css # Basic styles
# └──── script.js # Application scripts
## Launch
> **Move to the frontend folder**
> This is usually the `frontend` folder or similar.
> **Open the ``index.html`` file in your browser**
> You can simply double-click the file, or open it via the ``Open with'' context menu.
> **Recommended**
> Use a local server for fetch requests to work correctly (e.g. Live Server for VSCode or any http server).
> **Make sure that backend services are running**
## Usage
***Registration and Login***
- Enter your username and password.
- Click the “Register” link to register.
- After successful registration, log in.
***Enter Expression***
- Enter an expression (e.g., `2+2*3`) in the “Calculator” field.
- Click “Calculate.”
***View Results***
- Once the expression is submitted, it will appear in the list with a unique ID and status (pending or result).
- Click “Update Results” to update statuses.
***Clearing History***
- The “Clear All Calculations” button deletes all of the user's calculations.
## Example API request
* `POST /api/v1/login` - user login
* `POST /api/v1/register` - user registration
* `POST /submit-calculation` - send expression
* `GET /get-calculations-by-user?userId=...` - history of calculations
* `GET /get-calculation-result?id=...` - result by ID
* `POST /clear-all-calculations` - history clearing
* `GET /orchestrator-status` - orchestrator status
* `GET /ping-servers` - calculator statuses

Translated with DeepL.com (free version)
## Interface
![Illustration for the project]()