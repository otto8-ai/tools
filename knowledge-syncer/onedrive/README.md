# Knowledge OneDrive Sync

This project is a Go application that synchronizes files and folders from shared links using Microsoft Graph API. It reads metadata from a `metadata.json` file, processes the shared links, and downloads the files and folders to a specified working directory.

## Usage

1. Set the required environment variables:

   ```sh
   export GPTSCRIPT_GRAPH_MICROSOFT_COM_BEARER_TOKEN=<your-bearer-token>
   export GPTSCRIPT_WORKSPACE_DIR=<your-working-directory>
   ```

2. Create a `metadata.json` file in the working directory with the following structure:

   ```json
   {
     "input": {
       "sharedLinks": [
         "https://example.com/shared-link-1",
         "https://example.com/shared-link-2"
       ]
     }
   }
   ```

3. Run the application:

   ```sh
   gptscript github.com/gptscript-ai/knowledge-onedrive-sync
   ```

4. You will have output and files written into the working directory.

```json
{
  "output": {
    "status": "Done",
    "error": "",
    "files": {},
    "folders": {}
  }
}
```
