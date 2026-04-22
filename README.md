# X-Downloader

A lightweight media downloader for X (Twitter), built with Go.

## Usage Guide

### Prerequisites

- go 1.25 or higher

### Build

```bash
git clone https://github.com/we1sper/X-Downloader.git
cd X-Downloader/entry/x/cli && go build . -o cli
```

After the build is complete, an executable file named `cli` will be generated in the current directory.

### Run

#### 1. Generate template config file

Run the following command to generate a template config file in the current working directory:

```bash
./cli -t
```

The template content is as follows (required fields marked as `required`, optional fields as `optional`):

```json
{
  "Cookie": "required",
  "BearerToken": "required",
  "Retry": 3,
  "Downloader": 4,
  "Timeout": 60000,
  "SaveDir": "required",
  "Overwrite": false,
  "Delta": true,
  "Download": true,
  "Proxy": "optional",
  "LogLevel": "info",
  "LogFile": "optional",
  "BarrierCandidate": 10
}
```

The meaning of options:

| Option           | Functionality                                               | Required        |
|------------------|-------------------------------------------------------------|-----------------|
| Cookie           | X/Twitter account authentication cookie                     | ✅ Yes           |
| BearerToken      | X/Twitter API authentication Bearer Token                   | ✅ Yes           |
| Retry            | Number of retries for failed requests/downloads             | ❌ No            |
| Downloader       | Concurrent download goroutines                              | ❌ No            |
| Timeout          | Request/download timeout (ms)                               | ❌ No            |
| SaveDir          | Local directory for saving downloaded media files           | ✅ Yes           |
| Overwrite        | Overwrite existing files                                    | ❌ No            |
| Delta            | Enable incremental download mode (only fetch new content)   | ❌ No            |
| Download         | Auto-download media after fetching tweets                   | ❌ No            |
| Proxy            | HTTP proxy address (e.g., http://127.0.0.1:10808)           | ⚠️ Conditional* |
| LogLevel         | Log output level                                            | ❌ No            |
| LogFile          | Path to save log file (logs print to console only if empty) | ❌ No            |
| BarrierCandidate | Internal parameter (no modification needed)                 | ❌ No            |

> ⚠️ *Proxy Note: Required if your region cannot directly access X/Twitter; otherwise, optional.

#### 2. Download media tweets for a specific user

Run the following command to download media files (photos and videos) for the target user based on the config file:

```bash
./cli -u <username> -c <path/to/config>
```

Media files will be saved to the directory specified by the `SaveDir` field in the config file.
