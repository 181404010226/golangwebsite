# Golang Website Project

This project consists of a backend service built with Golang and uses various external services for authentication and data storage.

## Environment Configuration

The project requires several environment variables to be set up in a `.env` file. Below are the required configurations:

### Required Environment Variables

```env
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
SESSION_KEY=your_session_key
FRONTEND_URL=your_frontend_url
DATABASE_URI=your_mongodb_uri
HEROKU_API_KEY=your_heroku_api_key
```

### Environment Variables Description

- `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`: Required for GitHub OAuth authentication
- `SESSION_KEY`: Secret key for session management
- `FRONTEND_URL`: URL of the frontend application
- `DATABASE_URI`: MongoDB connection string
- `HEROKU_API_KEY`: API key for Heroku deployment

## Deployment

### Heroku Deployment

To deploy the application on Heroku, you need to set the following config vars:

```bash
heroku config:set GITHUB_CLIENT_ID=your_github_client_id
heroku config:set GITHUB_CLIENT_SECRET=your_github_client_secret
heroku config:set SESSION_KEY=your_session_key
heroku config:set DATABASE_URI=your_mongodb_uri
heroku config:set HEROKU_APP_NAME=your_app_name
heroku config:set FRONTEND_URL=your_frontend_url
```

## Security Notice

⚠️ **Important**: Never commit your `.env` file to version control. Add it to your `.gitignore` file to prevent accidentally exposing sensitive credentials.

## Getting Started

1. Clone the repository
2. Create a `.env` file in the root directory
3. Fill in the environment variables as described above
4. Run the application

## Tech Stack

- Backend: Golang
- Database: MongoDB
- Authentication: GitHub OAuth
- Deployment: Heroku 