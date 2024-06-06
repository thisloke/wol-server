#!/bin/bash

SERVER="delemaco"  # Server to ping

echo "Content-type: text/html"
echo
# Function to ping the server and generate HTML output
generate_html() {
    local server=$1

    # Ping the server (send 1 packet and wait for 1 second)
     ping -c 1 -W 1 "$server" > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        status="Online"
        color="green"
    else
        status="Booting"
        color="gray"
        wakeonlan b8:cb:29:a1:f3:88
    fi

    
    # Generate HTML content
    HTML_CONTENT=$(cat <<EOL
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Status</title>
    <style>
        body {
            background-color: $color;
            font-family: Arial, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            height: 100vh;
            margin: 0;
        }
        .status {
            font-size: 72px;
            color: white;
        }
    </style>
</head>
<body>
    <div class="status">
        Server <strong>$server</strong> is currently $status.
    </div>
</body>
</html>
EOL
)
    echo "$HTML_CONTENT"
}

# Run the function and save the output to a variable
HTML_OUTPUT=$(generate_html "$SERVER")

# Print the HTML content
echo "$HTML_OUTPUT"
