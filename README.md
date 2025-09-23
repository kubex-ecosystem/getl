# ![Getl Banner](./assets/getl_banner.png)

---

***Getl: An Efficient Data Synchronization and ETL Manager***

Getl is a powerful and flexible tool written in Go that streamlines the extraction, transformation, and loading (ETL) of data. It also supports continuous synchronization between various sources and destinations. Whether you're working with databases, file-based formats, or real-time messaging systems, Getl offers a unified approach to build dynamic data pipelines with ease.

---

## Table of Contents

1. [About the Project](#about-the-project)
2. [Features](#features)
3. [Installation](#installation)
4. [Usage](#usage)
   - [CLI Examples](#cli-examples)
   - [Configuration Examples](#configuration-examples)
5. [Configuration](#configuration)
6. [Roadmap](#roadmap)
7. [Contributing](#contributing)
8. [License](#license)
9. [Contact](#contact)

---

## About the Project

Getl is designed to be a robust solution for data integration and synchronization across heterogeneous systems. It supports a variety of data sources and destinations – from traditional relational databases to modern messaging systems – and bridges them through dynamic ETL workflows and flexible configurations.

**Why Getl?**

- **Ease of Use:** Configure and manage complex data flows with simple JSONC/YAML configuration files.
- **Flexibility:** Supports multiple databases and file formats with customizable field mappings and transformation rules.
- **Continuous Synchronization:** Schedule periodic syncs for near real-time updates.
- **Extensibility:** Integrate messaging systems such as Kafka or Redis for real-time data pipelines.

---

## Features

### 🚀 **NEW: Incremental Sync (Game Changer!)**

- **Intelligent Change Detection:** Only processes new or modified data, not entire datasets
- **Multiple Strategies:** Timestamp-based, Primary Key-based, and Hash-based sync
- **10x-300x Performance Boost:** Massive speed improvements for large datasets
- **State Management:** Automatic tracking of sync progress with recovery capabilities
- **Zero Configuration:** Works out-of-the-box with smart defaults

### 🔄 **Core ETL Features**

- **Data Extraction:** Connect and extract data from various sources (Oracle, PostgreSQL, MySQL, SQLite, SQL Server) using customizable SQL queries
- **Smart Type Detection:** Automatic data type inference and mapping between different database systems
- **Custom Transformations:** Define mapping and transformation rules to convert data between formats and structures
- **Data Loading and Synchronization:** Automatically create destination tables with proper types and constraints
- **Multiple Output Formats:** Export data to CSV, JSON, XML, YAML, PDF, and more
- **Dynamic Configuration:** Manage your ETL processes with configuration files that support comments (JSONC) for clarity

### ⚡ **Advanced Features**

- **Real-Time Integration:** Leverage messaging integrations for live data flow using Kafka, Redis, and others
- **Robust Error Handling:** Intelligent fallbacks and comprehensive logging
- **Production Ready:** Battle-tested type mapping and connection management
- **CLI Perfection:** Intuitive command-line interface with helpful output

---

## Installation

### Requirements

- **Go** (version 1.19 or above)
- Appropriate permissions to access data sources and destinations

```shell
# Clone this repository
git clone https://github.com/kubex-ecosystem/getl.git

# Navigate to the project directory
cd getl

# Build the binary using the Makefile
make build

# Install the binary (optional)
make install

# (Optional) Add the binary to your PATH
export PATH=$PATH:$(pwd)
```

---

## Usage

### CLI Examples

#### 🚀 **Incremental Sync (Recommended)**

```shell
# Sync only new orders (timestamp-based)
getl sync -f examples/configFiles/incremental_ecommerce_orders.json

# Sync only new leads (primary key-based)
getl sync -f examples/configFiles/incremental_crm_leads.json

# Sync analytics events with Kafka integration
getl sync -f examples/configFiles/incremental_analytics_events.json
```

#### 📊 **Traditional ETL**

```shell
# Basic synchronization: extracts data from a source and loads it into a destination
getl sync -f examples/configFiles/exp_config_a.json

# Extract data with a custom SQL query
getl extract --source "oracle_db" --query "SELECT * FROM products"

# Transform and load data using a custom configuration file
getl transform -f examples/configFiles/exp_config_b.json
```

### Configuration Examples

#### **🚀 NEW: Incremental Sync (High Performance)**

```json
{
  "sourceType": "postgres",
  "sourceConnectionString": "postgres://user:pass@localhost:5432/ecommerce",
  "sourceTable": "orders",
  "destinationType": "sqlite3",
  "destinationConnectionString": "/analytics/orders.db",
  "destinationTable": "orders_sync",
  "outputFormat": "csv",
  "outputPath": "/exports/orders_incremental.csv",

  "incrementalSync": {
    "enabled": true,
    "strategy": "timestamp",
    "timestampField": "created_at",
    "stateFile": "/state/orders-sync.json"
  }
}
```

> **📈 Performance**: Processes only new records, 10x-300x faster than full sync!

#### **Example 1: Basic Synchronization**

```json
{
  "sourceType": "godror",
  "sourceConnectionString": "username/password@127.0.0.1:1521/orcl",
  "destinationType": "sqlite3",
  "destinationConnectionString": "/home/user/.kubex/web/gorm.db",
  "destinationTable": "erp_products",
  "destinationTablePrimaryKey": "CODPARC",
  "sqlQuery": "SELECT P.CODPARC, P.NOMEPARC FROM TABLE P",
  "outputFormat": "csv",
  "outputPath": "/home/user/Documents/erp_products.csv",
  "needCheck": true,
  "checkMethod": "SELECT * FROM erp_products WHERE CODPARC = ? AND NOMEPARC = ?",
  "kafkaURL": "",
  "kafkaTopic": "",
  "kafkaGroupID": ""
}
```

#### **Example 2: Continuous Synchronization**

```json
{
  "sourceType": "godror",
  "sourceConnectionString": "username/password@127.0.0.1:1521/orcl",
  "destinationType": "sqlite3",
  "destinationConnectionString": "/home/user/.kubex/web/gorm.db",
  "destinationTable": "erp_products",
  "destinationTablePrimaryKey": "CODPARC",
  "sqlQuery": "SELECT P.CODPARC, P.NOMEPARC FROM TABLE P",
  "syncInterval": "30 * * * * *",
  "kafkaURL": "",
  "kafkaTopic": "",
  "kafkaGroupID": ""
}
```

#### **Example 3: Detailed Transformations**

```json
{
  "sourceType": "sqlite3",
  "sourceConnectionString": "/home/user/.kubex/web/gorm.db",
  "sourceTable": "erp_products",
  "sourceTablePrimaryKey": "CODPROD",
  "sqlQuery": "",
  "destinationType": "sqlServer",
  "destinationConnectionString": "sqlserver://username:password@localhost:1433?database=my_db_test&encrypt=disable&trustservercertificate=true",
  "destinationTable": "erp_products_test",
  "destinationTablePrimaryKey": "id_v",
  "kafkaURL": "",
  "kafkaTopic": "",
  "kafkaGroupID": "",
  "transformations": [
    {
      "sourceField": "CODPROD",
      "destinationField": "id_v",
      "operation": "none",
      "sPath": "erp_products",
      "dPath": "erp_products_test"
    },
    {
      "sourceField": "PRODDESCR",
      "destinationField": "name_v",
      "operation": "none",
      "sPath": "erp_products",
      "dPath": "erp_products_test"
    }
    // Additional transformations can be specified here.
  ]
}
```

---

## Configuration

Getl uses JSON or YAML configuration files (supporting JSONC for comments) to set up data source and destination connections, transformation rules, and synchronization intervals.

### 📚 **Documentation**

- **[Incremental Sync Guide](docs/INCREMENTAL_SYNC.md)** - Complete guide for high-performance incremental synchronization
- **[Configuration Examples](examples/configFiles/)** - Ready-to-use configuration files for common scenarios

### 🎯 **Key Configuration Sections**

- **Database Connections**: Support for Oracle, PostgreSQL, MySQL, SQLite, SQL Server
- **Incremental Sync**: Configure timestamp, primary key, or hash-based strategies
- **Transformations**: Field mapping and data transformation rules
- **Output Formats**: CSV, JSON, XML, YAML, PDF export options
- **Real-time Integration**: Kafka and messaging system configuration

---

## Roadmap

### ✅ **Recently Completed:**

- **Incremental Sync Engine**: Timestamp, Primary Key, and Hash-based strategies
- **Smart Type Detection**: Automatic data type inference and mapping
- **State Management**: Robust sync state tracking and recovery
- **Production-Ready CLI**: Intuitive interface with comprehensive logging

### 🔜 **Coming Soon:**

- **Hash-based Sync**: Complete implementation for change detection
- **Bidirectional Sync**: Two-way data synchronization
- **Auto-discovery**: Automatic schema and relationship detection
- **Web Dashboard**: Real-time monitoring and management interface
- **Advanced Transformations**: Custom functions and complex data operations
- **Enhanced Kafka Integration**: Stream processing and real-time pipelines

---

## Contributing

Contributions are welcome!
Please feel free to open issues or submit pull requests. For more details, see the [Contributing Guide](CONTRIBUTING.md).

---

## License

This project is licensed under the [MIT License](LICENSE).

---

## Contact

- **Developer:** [Rafael Mori](mailto:faelmori@gmail.com)
- **GitHub:** [faelmori](https://github.com/faelmori)

If you find this project interesting or would like to collaborate, please reach out!

---
