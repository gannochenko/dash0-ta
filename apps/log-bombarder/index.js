import grpc from "k6/net/grpc";
import { check } from "k6";
import {
  randomString,
  randomIntBetween,
} from "https://jslib.k6.io/k6-utils/1.2.0/index.js";

const client = new grpc.Client();
client.load(["proto"], "logs_service.proto");

export const options = {
  vus: 10, // 10 virtual users
  duration: "30s", // run for 30 seconds
};

// Helper function to generate random hex string
function randomHex(length) {
  const chars = "0123456789ABCDEF";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return result;
}

// Helper function to generate trace ID (32 hex chars)
function generateTraceId() {
  return randomHex(32);
}

// Helper function to generate span ID (16 hex chars)
function generateSpanId() {
  return randomHex(16);
}

// Helper function to convert hex string to Uint8Array
function hexToBytes(hex) {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substr(i, 2), 16);
  }
  return bytes;
}

// Generate random log data in the format specified by the user
function generateRandomLogData() {
  const serviceNames = [
    "my.service",
    "api.service",
    "web.service",
    "db.service",
  ];
  const severityLevels = [
    { number: 5, text: "Debug" },
    { number: 9, text: "Info" },
    { number: 10, text: "Information" },
    { number: 13, text: "Warn" },
    { number: 17, text: "Error" },
  ];

  const severity =
    severityLevels[Math.floor(Math.random() * severityLevels.length)];
  const now = Date.now() * 1000000; // Convert to nanoseconds

  return {
    resource_logs: [
      {
        resource: {
          attributes: [
            {
              key: "service.name",
              value: {
                string_value:
                  serviceNames[Math.floor(Math.random() * serviceNames.length)],
              },
            },
          ],
        },
        scope_logs: [
          {
            scope: {
              name: "my.library",
              version: "1.0.0",
              attributes: [
                {
                  key: "my.scope.attribute",
                  value: {
                    string_value: "some scope attribute",
                  },
                },
              ],
            },
            log_records: [
              {
                time_unix_nano: now.toString(),
                observed_time_unix_nano: now.toString(),
                severity_number: severity.number,
                severity_text: severity.text,
                trace_id: "5B8EFFF798038103D269B633813FC60C", //hexToBytes(generateTraceId()),
                span_id: "0000000000000000", //hexToBytes(generateSpanId()),
                body: {
                  string_value: `Example log record ${randomString(8)}`,
                },
                attributes: [
                  {
                    key: "string.attribute",
                    value: {
                      string_value: randomString(10),
                    },
                  },
                  {
                    key: "boolean.attribute",
                    value: {
                      bool_value: Math.random() > 0.5,
                    },
                  },
                  {
                    key: "int.attribute",
                    value: {
                      int_value: randomIntBetween(1, 100).toString(),
                    },
                  },
                  {
                    key: "double.attribute",
                    value: {
                      double_value: Math.random() * 1000,
                    },
                  },
                  {
                    key: "array.attribute",
                    value: {
                      array_value: {
                        values: [
                          {
                            string_value: randomString(5),
                          },
                          {
                            string_value: randomString(7),
                          },
                        ],
                      },
                    },
                  },
                  {
                    key: "map.attribute",
                    value: {
                      kvlist_value: {
                        values: [
                          {
                            key: "some.map.key",
                            value: {
                              string_value: randomString(12),
                            },
                          },
                        ],
                      },
                    },
                  },
                ],
              },
            ],
          },
        ],
      },
    ],
  };
}

export default function () {
  client.connect("localhost:443", { plaintext: true, reflect: true });

  const data = generateRandomLogData();

  const response = client.invoke(
    "opentelemetry.proto.collector.logs.v1.LogsService/Export",
    data
  );

  check(response, {
    "status is OK": (r) => r && r.status === grpc.StatusOK,
  });

  client.close();
}
