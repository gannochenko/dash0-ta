# Log Bombarder

A k6-based load testing tool that bombards the gRPC log processor service with random OpenTelemetry log data.

## Prerequisites

1. Install k6: https://k6.io/docs/getting-started/installation/

   ```bash
   # On macOS with Homebrew
   brew install k6

   # Or use the npm script
   npm run install-k6
   ```

2. Make sure the log processor service is running on `localhost:443`

## Usage

### Basic Run

```bash
k6 run bombarder.js
```

### Custom Configuration

You can modify the test configuration by editing the `options` object in `bombarder.js`:

```javascript
export const options = {
  vus: 10, // Number of virtual users (concurrent connections)
  duration: "30s", // Test duration
};
```

### Advanced Options

```bash
# Run with custom VUs and duration
k6 run --vus 50 --duration 60s bombarder.js

# Run with stages (ramping up/down)
k6 run --stage 10s:10,30s:50,10s:0 bombarder.js
```

## What it does

The bombarder generates random OpenTelemetry log data that matches the format specified in the requirements:

- **Resource attributes**: Random service names
- **Scope information**: Fixed library name and version with scope attributes
- **Log records**: Random severity levels, timestamps, trace/span IDs, and various attribute types:
  - String attributes
  - Boolean attributes
  - Integer attributes
  - Double attributes
  - Array attributes
  - Map/KVList attributes

Each virtual user will:

1. Connect to the gRPC service
2. Generate random log data
3. Send it via the `Export` method
4. Verify the response status
5. Close the connection
6. Repeat for the test duration

## Output

k6 will display real-time metrics including:

- Request rate (RPS)
- Response times
- Success/failure rates
- Data transfer rates

## Customization

- **Service endpoint**: Change `localhost:443` in the `client.connect()` call
- **Data generation**: Modify the `generateRandomLogData()` function
- **Load pattern**: Adjust the `options` configuration
- **Service names**: Edit the `serviceNames` array
- **Severity levels**: Modify the `severityLevels` array
