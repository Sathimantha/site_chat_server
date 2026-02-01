# /etc/systemd/system/geminichatserver.service

[Unit]
Description=Gemini Chat Server Application
After=network.target

[Service]
Type=simple
ExecStart=/home/bitnami/work/site_chat_server/server
WorkingDirectory=/home/bitnami/work/site_chat_server
Restart=on-failure
RestartSec=5
User=bitnami

# Environment variables ##############################
# OpenAI API key
Environment=GOOGLE_API_KEY=#your_google_api_key_here#

# Server port
Environment=PORT=5004

# SSL certificate and key paths
Environment=SSL_CERT_PATH='/home/bitnami/work/site_chat_server/certs/server.crt'
Environment=SSL_KEY_PATH='/home/bitnami/work/site_chat_server/certs/server.key'


# CORS allowed origins
Environment=ALLOWED_ORIGINS=https://*.peaceandhumanity.org,https://peaceandhumanity.org,https://preview.peaceandhumanity.org

# System prompt
Environment=SYSTEM_PROMPT_PATH='/home/bitnami/work/site_chat_server/system_prompt.md'

# Database file path (relative or absolute)
Environment=DB_PATH='/home/bitnami/work/site_chat_server/chats.db'

######################################################

KillMode=process
TimeoutStartSec=30
TimeoutStopSec=10

[Install]
WantedBy=multi-user.target


