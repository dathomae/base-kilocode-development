# fastparquet Overview

## Introduction

fastparquet is a Python interface to the Parquet file format. The [Parquet format](https://github.com/apache/parquet-format) is a common binary data store, used particularly in the Hadoop/big-data sphere. It provides several advantages relevant to big-data processing:

1. **Columnar storage** - Only read the data of interest
2. **Efficient binary packing** - Compact storage format
3. **Choice of compression algorithms and encoding** - Flexibility for different use cases
4. **Split data into files** - Allows for parallel processing
5. **Range of logical types** - Support for various data types
6. **Statistics stored in metadata** - Allows for skipping unneeded chunks
7. **Data partitioning using the directory structure** - Organized data storage

Since Parquet was developed as part of the Hadoop ecosystem, its reference implementation is written in Java. fastparquet provides a performant library to read and write Parquet files from Python, without any need for a Python-Java bridge. This makes the Parquet format an ideal storage mechanism for Python-based big data workflows.

The tabular nature of Parquet is a good fit for Pandas data-frame objects, and fastparquet exclusively deals with data-frame↔Parquet conversion.

## Key Features

1. **Read and write Parquet files** - In single or multiple-file format
2. **Compression options** - Choice of compression per-column and various optimized encoding schemes; ability to choose row divisions and partitioning on write
3. **Performance** - Acceleration of both reading and writing using Cython, competitive performance versus other frameworks
4. **File-like object support** - Ability to read and write to arbitrary file-like objects, allowing interoperability with `fsspec` filesystems and others
5. **Dask integration** - Can be called from [dask](http://dask.pydata.org) to enable parallel reading and writing with Parquet files, possibly distributed across a cluster

## Relation to Other Projects

| Project                                                              | Description                                                                                                                                                    |
| -------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [parquet-python](https://github.com/jcrobak/parquet-python)          | The original pure-Python Parquet quick-look utility which was the inspiration for fastparquet                                                                  |
| Apache [Arrow](http://pyarrow.readthedocs.io/en/latest/)             | Defines an in-memory data representation and can read/write parquet, including conversion to pandas. It is the "other" engine available within Dask and Pandas |
| [PySpark](http://Spark.apache.org/docs/2.1.0/programming-guide.html) | A Python API to the Spark engine that gives Python programmers access to the Parquet format                                                                    |
| [dask](http://dask.pydata.org)                                       | fastparquet lives within the dask ecosystem and is designed to work well with dask for parallel execution                                                      |

## Installation

Install fastparquet using pip or conda:

```bash
# Using pip
pip install fastparquet

# Using conda
conda install -c conda-forge fastparquet
```

### Requirements

- Python 3.8+
- pandas
- numpy
- cramjam (for compression)

## Documentation Reference

For the complete and most up-to-date documentation, see: https://fastparquet.readthedocs.io/en/latest/
