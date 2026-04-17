# Data Pipelines

## Diagrams

### Fluent Bit

```mermaid
flowchart LR

  classDef input    font-weight:bold,fill:#F0A500,color:#000,stroke:#C07800,stroke-width:2px
  classDef filter   font-weight:bold,fill:#4A90D9,color:#fff,stroke:#2C5F8A,stroke-width:2px
  classDef output   font-weight:bold,fill:#50C878,color:#000,stroke:#2E7A47,stroke-width:2px
  classDef backend  font-weight:bold,fill:#E8608A,color:#fff,stroke:#A03058,stroke-width:2px
  classDef frontend font-weight:bold,fill:#9B6DD4,color:#fff,stroke:#6A3FA0,stroke-width:2px

  subgraph PIPELINES
    direction TB

    subgraph LOGS["Forward Logs"]
      direction LR

      LI1(["INPUT::File"]):::input
      LI2(["INPUT::Forward"]):::input
      LP1("FILTER::modify"):::filter
      LP2("FILTER::grep"):::filter
      LO1(["OUTPUT::Stdout"]):::output
      LO2(["OUTPUT::File"]):::output
      LO3(["OUTPUT::Forward"]):::output
      LO4(["OUTPUT::Loki"]):::output

      LI1 --> LP1
      LI2 --> LP1
      LP1 --> LP2
      LP2 --> LO1
      LP2 --> LO2
      LP2 --> LO3
      LP2 --> LO4
    end

    subgraph METRICS["Prometheus Metrics"]
      direction LR

      MI1(["INPUT::Scrape /metrics"]):::input
      MI2(["INPUT::Prometheus Remote Write"]):::input
      MP1("FILTER::modify"):::filter
      MP2("FILTER::grep"):::filter
      MO1(["OUTPUT::Stdout"]):::output
      MO2(["OUTPUT::File"]):::output
      MO3(["OUTPUT::Expose /metrics"]):::output
      MO4(["OUTPUT::Prometheus Remote Write"]):::output

      MI1 --> MP1
      MI2 --> MP1
      MP1 --> MP2
      MP2 --> MO1
      MP2 --> MO2
      MP2 --> MO3
      MP2 --> MO4
    end

    subgraph OTEL["OpenTelemetry"]
      direction LR

      OI1(["INPUT::OTLP HTTP/gRPC"]):::input
      OP1("FILTER::modify"):::filter
      OP2("FILTER::grep"):::filter
      OO1(["OUTPUT::Stdout"]):::output
      OO2(["OUTPUT::File"]):::output
      OO3(["OUTPUT::OTLP HTTP"]):::output

      OI1 --> OP1
      OP1 --> OP2
      OP2 --> OO1
      OP2 --> OO2
      OP2 --> OO3
    end
  end

  subgraph VOLUMES
    direction TB

    V1[("fluentbit_data")]:::backend

    LO2 --> V1
    MO2 --> V1
    OO2 --> V1
  end

  subgraph BACKENDS
    direction TB

    B1[("Loki")]:::backend
    B2[("Mimir")]:::backend
    B3[("Tempo")]:::backend

    LO4 --> B1
    MO4 --> B2
    OO3 --> B1
    OO3 --> B2
    OO3 --> B3
  end

  subgraph FRONTENDS
    direction TB

    F1(("Grafana")):::frontend

    B1 --> F1
    B2 --> F1
    B3 --> F1
  end
```

### OpenTelemetry Collector

```mermaid
flowchart LR

  classDef receiver  font-weight:bold,fill:#F0A500,color:#000,stroke:#C07800,stroke-width:2px
  classDef processor font-weight:bold,fill:#4A90D9,color:#fff,stroke:#2C5F8A,stroke-width:2px
  classDef exporter  font-weight:bold,fill:#50C878,color:#000,stroke:#2E7A47,stroke-width:2px
  classDef backend   font-weight:bold,fill:#E8608A,color:#fff,stroke:#A03058,stroke-width:2px
  classDef frontend  font-weight:bold,fill:#9B6DD4,color:#fff,stroke:#6A3FA0,stroke-width:2px

  subgraph PIPELINES
    direction TB

    subgraph LOGS["OpenTelemetry Logs"]
      direction LR

      LR1(["RECEIVER::File"]):::receiver
      LR2(["RECEIVER::Forward"]):::receiver
      LR3(["RECEIVER::OTLP HTTP"]):::receiver
      LR4(["RECEIVER::OTLP gRPC"]):::receiver
      LP1("PROCESSOR::memory_limiter"):::processor
      LP2("PROCESSOR::resourcedetection"):::processor
      LP3("PROCESSOR::resource"):::processor
      LP4("PROCESSOR::transform"):::processor
      LP5("PROCESSOR::filter"):::processor
      LP6("PROCESSOR::batch"):::processor
      LE1(["EXPORTER::Stdout"]):::exporter
      LE2(["EXPORTER::File"]):::exporter
      LE3(["EXPORTER::OTLP HTTP"]):::exporter
      LE4(["EXPORTER::OTLP gRPC"]):::exporter

      LR1 --> LP1
      LR2 --> LP1
      LR3 --> LP1
      LR4 --> LP1
      LP1 --> LP2 --> LP3 --> LP4 --> LP5 --> LP6
      LP6 --> LE1
      LP6 --> LE2
      LP6 --> LE3
      LP6 --> LE4
    end

    subgraph METRICS["OpenTelemetry Metrics"]
      direction LR

      MR1(["RECEIVER::Scrape /metrics"]):::receiver
      MR2(["RECEIVER::Prometheus Remote Write"]):::receiver
      MR3(["RECEIVER::OTLP HTTP"]):::receiver
      MR4(["RECEIVER::OTLP gRPC"]):::receiver
      MP1("PROCESSOR::memory_limiter"):::processor
      MP2("PROCESSOR::resourcedetection"):::processor
      MP3("PROCESSOR::resource"):::processor
      MP4("PROCESSOR::transform"):::processor
      MP5("PROCESSOR::filter"):::processor
      MP6("PROCESSOR::batch"):::processor
      ME1(["EXPORTER::Stdout"]):::exporter
      ME2(["EXPORTER::File"]):::exporter
      ME3(["EXPORTER::Expose /metrics"]):::exporter
      ME4(["EXPORTER::Prometheus Remote Write"]):::exporter
      ME5(["EXPORTER::OTLP HTTP"]):::exporter
      ME6(["EXPORTER::OTLP gRPC"]):::exporter

      MR1 --> MP1
      MR2 --> MP1
      MR3 --> MP1
      MR4 --> MP1
      MP1 --> MP2 --> MP3 --> MP4 --> MP5 --> MP6
      MP6 --> ME1
      MP6 --> ME2
      MP6 --> ME3
      MP6 --> ME4
      MP6 --> ME5
      MP6 --> ME6
    end

    subgraph TRACES["OpenTelemetry Traces"]
      direction LR

      TR1(["RECEIVER::OTLP HTTP"]):::receiver
      TR2(["RECEIVER::OTLP gRPC"]):::receiver
      TP1("PROCESSOR::memory_limiter"):::processor
      TP2("PROCESSOR::resourcedetection"):::processor
      TP3("PROCESSOR::resource"):::processor
      TP4("PROCESSOR::transform"):::processor
      TP5("PROCESSOR::filter"):::processor
      TP6("PROCESSOR::batch"):::processor
      TE1(["EXPORTER::Stdout"]):::exporter
      TE2(["EXPORTER::File"]):::exporter
      TE3(["EXPORTER::OTLP HTTP"]):::exporter
      TE4(["EXPORTER::OTLP gRPC"]):::exporter

      TR1 --> TP1
      TR2 --> TP1
      TP1 --> TP2 --> TP3 --> TP4 --> TP5 --> TP6
      TP6 --> TE1
      TP6 --> TE2
      TP6 --> TE3
      TP6 --> TE4
    end
  end

  subgraph VOLUMES
    direction TB

    V1[("opentelemetry_data")]:::backend

    LE2 --> V1
    ME2 --> V1
    TE2 --> V1
  end

  subgraph BACKENDS
    direction TB

    B1[("Loki")]:::backend
    B2[("Mimir")]:::backend
    B3[("Tempo")]:::backend

    LE3 --> B1
    ME4 --> B2
    TE4 --> B3
  end

  subgraph FRONTENDS
    direction TB

    F1(("Grafana")):::frontend

    B1 --> F1
    B2 --> F1
    B3 --> F1
  end
```

### Grafana Alloy

```mermaid
flowchart LR

  classDef receiver  font-weight:bold,fill:#F0A500,color:#000,stroke:#C07800,stroke-width:2px
  classDef processor font-weight:bold,fill:#4A90D9,color:#fff,stroke:#2C5F8A,stroke-width:2px
  classDef exporter  font-weight:bold,fill:#50C878,color:#000,stroke:#2E7A47,stroke-width:2px
  classDef backend   font-weight:bold,fill:#E8608A,color:#fff,stroke:#A03058,stroke-width:2px
  classDef frontend  font-weight:bold,fill:#9B6DD4,color:#fff,stroke:#6A3FA0,stroke-width:2px

  subgraph PIPELINES
    direction TB

    subgraph LOGS["Loki Logs"]
      direction LR

      LR1(["COMPONENT::File"]):::receiver
      LR2(["COMPONENT::Loki Push API"]):::receiver
      LP1("COMPONENT::relabel"):::processor
      LP2("COMPONENT::process"):::processor
      LE1(["COMPONENT::Stdout"]):::exporter
      LE2(["COMPONENT::Loki Push API"]):::exporter

      LR1 --> LP1
      LR2 --> LP1
      LP1 --> LP2
      LP2 --> LE1
      LP2 --> LE2
    end

    subgraph METRICS["Prometheus Metrics"]
      direction LR

      MR1(["COMPONENT::Scrape /metrics"]):::receiver
      MR2(["COMPONENT::Prometheus Remote Write"]):::receiver
      MP1("COMPONENT::relabel"):::processor
      ME1(["COMPONENT::Stdout"]):::exporter
      ME2(["COMPONENT::Prometheus Remote Write"]):::exporter

      MR1 --> MP1
      MR2 --> MP1
      MP1 --> ME1
      MP1 --> ME2
    end

    subgraph PPROFS["Performance Profiles"]
      direction LR

      PR1(["COMPONENT::Scrape /debug/pprof"]):::receiver
      PR2(["COMPONENT::Pyroscope Ingestion API"]):::receiver
      PP1("COMPONENT::relabel"):::processor
      PE1(["COMPONENT::Pyroscope Ingestion API"]):::exporter

      PR1 --> PP1
      PR2 --> PP1
      PP1 --> PE1
    end

    subgraph OTEL_LOGS["OpenTelemetry Logs"]
      direction LR

      OLR1(["RECEIVER::File"]):::receiver
      OLR2(["RECEIVER::Forward"]):::receiver
      OLR3(["RECEIVER::OTLP HTTP"]):::receiver
      OLR4(["RECEIVER::OTLP gRPC"]):::receiver
      OLP1("PROCESSOR::memory_limiter"):::processor
      OLP2("PROCESSOR::resourcedetection"):::processor
      OLP3("PROCESSOR::resource"):::processor
      OLP4("PROCESSOR::transform"):::processor
      OLP5("PROCESSOR::filter"):::processor
      OLP6("PROCESSOR::batch"):::processor
      OLE1(["EXPORTER::Stdout"]):::exporter
      OLE2(["EXPORTER::File"]):::exporter
      OLE3(["EXPORTER::Loki Push API"]):::exporter
      OLE4(["EXPORTER::OTLP HTTP"]):::exporter
      OLE5(["EXPORTER::OTLP gRPC"]):::exporter

      OLR1 --> OLP1
      OLR2 --> OLP1
      OLR3 --> OLP1
      OLR4 --> OLP1
      OLP1 --> OLP2 --> OLP3 --> OLP4 --> OLP5 --> OLP6
      OLP6 --> OLE1
      OLP6 --> OLE2
      OLP6 --> OLE3
      OLP6 --> OLE4
      OLP6 --> OLE5
    end

    subgraph OTEL_METRICS["OpenTelemetry Metrics"]
      direction LR

      OMR1(["RECEIVER::OTLP HTTP"]):::receiver
      OMR2(["RECEIVER::OTLP gRPC"]):::receiver
      OMP1("PROCESSOR::memory_limiter"):::processor
      OMP2("PROCESSOR::resourcedetection"):::processor
      OMP3("PROCESSOR::resource"):::processor
      OMP4("PROCESSOR::transform"):::processor
      OMP5("PROCESSOR::filter"):::processor
      OMP6("PROCESSOR::batch"):::processor
      OME1(["EXPORTER::Stdout"]):::exporter
      OME2(["EXPORTER::File"]):::exporter
      OME3(["EXPORTER::Prometheus Remote Write"]):::exporter
      OME4(["EXPORTER::OTLP HTTP"]):::exporter
      OME5(["EXPORTER::OTLP gRPC"]):::exporter

      OMR1 --> OMP1
      OMR2 --> OMP1
      OMP1 --> OMP2 --> OMP3 --> OMP4 --> OMP5 --> OMP6
      OMP6 --> OME1
      OMP6 --> OME2
      OMP6 --> OME3
      OMP6 --> OME4
      OMP6 --> OME5
    end

    subgraph OTEL_TRACES["OpenTelemetry Traces"]
      direction LR

      OTR1(["RECEIVER::OTLP HTTP"]):::receiver
      OTR2(["RECEIVER::OTLP gRPC"]):::receiver
      OTP1("PROCESSOR::memory_limiter"):::processor
      OTP2("PROCESSOR::resourcedetection"):::processor
      OTP3("PROCESSOR::resource"):::processor
      OTP4("PROCESSOR::transform"):::processor
      OTP5("PROCESSOR::filter"):::processor
      OTP6("PROCESSOR::batch"):::processor
      OTE1(["EXPORTER::Stdout"]):::exporter
      OTE2(["EXPORTER::File"]):::exporter
      OTE3(["EXPORTER::OTLP HTTP"]):::exporter
      OTE4(["EXPORTER::OTLP gRPC"]):::exporter

      OTR1 --> OTP1
      OTR2 --> OTP1
      OTP1 --> OTP2 --> OTP3 --> OTP4 --> OTP5 --> OTP6
      OTP6 --> OTE1
      OTP6 --> OTE2
      OTP6 --> OTE3
      OTP6 --> OTE4
    end
  end

  subgraph VOLUMES
    direction TB

    V1[("alloy_data")]:::backend

    OLE2 --> V1
    OME2 --> V1
    OTE2 --> V1
  end

  subgraph BACKENDS
    direction TB

    B1[("Loki")]:::backend
    B2[("Mimir")]:::backend
    B3[("Tempo")]:::backend
    B4[("Pyroscope")]:::backend

    LE2  --> B1
    ME2  --> B2
    OLE3 --> B1
    OME3 --> B2
    OTE4 --> B3
    PE1  --> B4
  end

  subgraph FRONTENDS
    direction TB

    F1(("Grafana")):::frontend

    B1 --> F1
    B2 --> F1
    B3 --> F1
    B4 --> F1
  end
```

## Considerations

### Scraping Telemetry Data

In this setup, all scraping configurations for metrics and performance profiling data are centralized in Alloy.
Alloy acts as the single collector responsible for scraping `/metrics` and `/debug/pprof` endpoints across all services,
then shipping the collected data to the respective storage backends (*Mimir* for metrics, *Pyroscope* for profiles).

In a real-world production environment, this centralized approach will not scale well.
Depending on your telemetry data pipeline architecture and the number of services you operate,
distributing the collection workload across multiple collector instances is strongly recommended.

A common pattern for distributed collection is the **sidecar architecture**,
where a dedicated collector instance (e.g., Alloy) is deployed alongside each container within every pod.
Each sidecar scrapes metrics and pprof data exclusively from `localhost`,
ensuring it only collects telemetry from the co-located service.
The data is then forwarded to a central aggregator or shipped directly to the storage backends.

## Resources

  - **Mermaid**
    - [Syntax and Configuration](https://mermaid.js.org/intro/syntax-reference.html)
    - [Flowchart](https://mermaid.js.org/syntax/flowchart.html)
