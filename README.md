![Getl Banner](./assets/getl_banner.png)

---

**Getl: An Efficient Data Synchronization and ETL Manager**

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
- **Data Extraction:** Connect and extract data from various sources (Oracle, PostgreSQL, MySQL, SQLite, MongoDB, etc.) using customizable SQL queries.
- **Custom Transformations:** Define mapping and transformation rules to convert data between formats and structures.
- **Data Loading and Synchronization:** Automatically create destination tables (with validations and constraints) and load data efficiently; supports both batch and continuous sync modes.
- **Multiple Output Formats:** Export data to CSV, JSON, XML, YAML, PDF, and more.
- **Dynamic Configuration:** Manage your ETL processes with configuration files that support comments (JSONC) for clarity.
- **Real-Time Integration:** Leverage messaging integrations for live data flow using Kafka, Redis, and others.

---

## Installation

### Requirements
- **Go** (version 1.19 or above)
- Appropriate permissions to access data sources and destinations

```shell
# Clone this repository
git clone https://github.com/faelmori/getl.git

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
Below are some examples of how to use Getl’s CLI:

```shell
# Basic synchronization: extracts data from a source and loads it into a destination
getl sync -f examples/configFiles/exp_config_a.json

# Extract data with a custom SQL query
getl extract --source "oracle_db" --query "SELECT * FROM products"

# Transform and load data using a custom configuration file
getl transform -f examples/configFiles/exp_config_b.json
```

### Configuration Examples

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
Getl uses JSON or YAML configuration files (supporting JSONC for comments) to set up data source and destination connections, transformation rules, and synchronization intervals. These files are central to configuring the ETL process, and detailed documentation is available in the [Configuration Documentation](https://github.com/faelmori/getl/README.md#configuration-file).

---

## Roadmap
🔜 **Planned Features:**
- Support for additional data sources and destinations (e.g., MongoDB, Redis)
- Enhanced transformation operations and custom processing functions
- Expanded real-time data integration via Kafka and Redis
- A dashboard for monitoring the status and performance of ETL jobs

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