# Content Automation Pipeline

A background Go worker that automatically fetches trending tech news, scores them based on popularity and recency, uses AI (OpenAI, Claude, or Gemini) to rewrite them into engaging social media posts, and pushes them directly to a Notion Database as a publishing queue.

## 🌟 Features

- **Automated Collection**: Fetches daily content from HackerNews and Dev.to.
- **Intelligent Scoring**: Ranks articles based on recency, popularity, and keyword relevance.
- **AI Rewrites**: Automatically generates the post body, relevant hashtags, and a suggested image prompt using your choice of LLM (OpenAI, Claude, or Gemini).
- **Notion Integration**: Directly queues the top 10 generated posts into a Notion database.
- **Cron Scheduling**: Built-in `cron` scheduling to run the pipeline automatically every day at 8:00 AM.

## 🏗 Architecture

- **`internal/collector`**: API fetchers for data sources.
- **`internal/filter`**: The scoring and ranking engine.
- **`internal/generator`**: The AI factory that routes to your provided LLM.
- **`internal/publisher`**: The Notion API wrapper to create database pages.
- **`internal/scheduler`**: The orchestrator that ties the entire pipeline together.

## 🚀 Getting Started

### Prerequisites
- Go 1.21+
- A Notion Integration Token & Database ID
- An API Key for OpenAI, Anthropic, or Google Gemini.

### 1. Configure the Environment
Create a `.env` file in the root directory (you can copy `.env.example`):
```bash
cp .env.example .env
```
Fill in your API keys and Notion details.

### 2. Set up Notion Database
Ensure your Notion database has the following properties:
- `Title` (Title)
- `Category` (Select)
- `Source` (URL)
- `Generated Post` (Text)
- `Hashtags` (Text)
- `Image Prompt` (Text)
- `Status` (Select)
- `Date` (Date)

**Critical:** Make sure to click the `...` menu on your database page, go to "Connections", and explicitly share the database with your Notion Integration!

### 3. Run the Pipeline
Start the background worker:
```bash
go run ./cmd/worker
```

When `ENV` is not set to `production`, the worker will run every minute for testing. In `production`, it runs daily at 8:00 AM.

## 📜 Agent Guidelines
If you are an AI assistant working on this repository, please read the `AGENTS.md` and `CLAUDE.md` files for structural rules and coding guidelines.


To run your Go pipeline in the background on an AWS EC2 Linux server reliably, the best approach is to create a systemd service. This ensures your application automatically restarts if it crashes and automatically starts when the EC2 instance reboots.

Here is a step-by-step guide to setting it up:

1. Build the Go Binary
First, compile your Go code into a single executable binary on your EC2 server. Run this inside your project directory:

```bash
# Compile the worker with optimizations to reduce binary size
go build -ldflags="-s -w" -trimpath -o content-pipeline ./cmd/worker
```

2. Create the Systemd Service File
Create a new service configuration file in the systemd directory using your preferred text editor (like nano):

```bash
sudo nano /etc/systemd/system/content-pipeline.service
```

Add the following configuration. Make sure to replace `/path/to/your/project` with the actual absolute path to your project folder (e.g., `/home/ubuntu/content-automation-pipeline` or `/home/ec2-user/...`):

```ini
[Unit]
Description=Content Automation Pipeline Worker
After=network.target

[Service]
# Change this to your EC2 user (e.g., ubuntu for Ubuntu AMIs, ec2-user for Amazon Linux)
User=ubuntu
Group=ubuntu

# Set the working directory where your .env file is located
WorkingDirectory=/path/to/your/project

# Prevent Out-Of-Memory (OOM) crashes on 1GB RAM instances
Environment=GOMEMLIMIT=250MiB

# Command to execute the binary
ExecStart=/path/to/your/project/content-pipeline

# Restart automatically if the app crashes
Restart=always
RestartSec=5

# Optional: Ensure it picks up environment variables if you don't use the .env file
# Environment=ENV=production

[Install]
WantedBy=multi-user.target
```
Save and exit (Ctrl+O, Enter, Ctrl+X).

3. Enable and Start the Service
Now, reload systemd to recognize your new service, enable it to start on boot, and start it immediately:

bash
# Reload systemd manager configuration
sudo systemctl daemon-reload
# Enable the service to start automatically on system boot
sudo systemctl enable content-pipeline
# Start the service now
sudo systemctl start content-pipeline
4. Check the Status and Logs
You can verify that it is running successfully in the background:

bash
# Check the current status
sudo systemctl status content-pipeline
Because you are using structured logging (zap), your application logs will be perfectly captured by journalctl. You can view real-time logs of your background worker by running:

bash
# Follow the live logs
sudo journalctl -u content-pipeline -f
Alternative (Quick & Dirty method): If you just want to run it quickly without setting up a service, you can use nohup. It runs the process ignoring hangup signals and redirects output to a file:

bash
nohup go run ./cmd/worker > pipeline.log 2>&1 &
(You can see your logs by running tail -f pipeline.log)