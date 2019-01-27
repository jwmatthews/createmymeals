# createmymeals
Curate recipes and help with weekly meal planning

## Notes on talking to GMail to parse messages to obtain recipe links
1. Download 'credentials.json' from enabling the Gmail API
    * Follow 'Turn on the Gmail API' at: https://developers.google.com/gmail/api/quickstart/go
    * Copy 'credentials.json' to the 'build' directory
2. `make build`
3. `cd build && ./list_recipes`
4. Running the binary will display a URL you must visit in your web browser to obtain a token from Gmail API.
    * Visit the displayed URL from the terminal
    * Obtain the token to talk to your gmail account
    * The binary will store the token to disk as 'token.json' in the 'build' directory
        * Ensure you don't check in 'token.json' and 'credentials.json'
        * Note: 'build/.gitignore' has entries to ignore 'token.json' and 'credentials.json'