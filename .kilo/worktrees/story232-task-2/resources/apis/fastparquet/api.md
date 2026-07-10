# fastparquet API Reference

## Main Functions and Classes

| Function/Class                                                           | Description                                            |
| ------------------------------------------------------------------------ | ------------------------------------------------------ |
| [`ParquetFile`](#parquetfile-class)                                      | The metadata of a parquet file or collection           |
| [`write()`](#write-function)                                             | Write pandas dataframe to filename with parquet format |
| [`update_file_custom_metadata()`](#update_file_custom_metadata-function) | Update custom metadata in a parquet file               |

---

## ParquetFile Class

```python
class fastparquet.ParquetFile(fn, verify=False, open_with=<built-in function open>,
                               root=False, sep=None, fs=None, pandas_nulls=True,
                               dtypes=None)
```

The metadata of a parquet file or collection. Reads the metadata (row-groups and schema definition) and provides methods to extract the data from the files.

Note that when reading parquet files partitioned using directories (i.e. using the hive/drill scheme), an attempt is made to coerce the partition values to a number, datetime or timedelta. Fastparquet cannot read a hive/drill parquet file with partition names which coerce to the same value, such as "0.7" and ".7".

### Parameters

| Parameter      | Type                             | Description                                                                                                                                                                                                                                    |
| -------------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `fn`           | path/URL string or list of paths | Location of the data. If a directory, will attempt to read a file "\_metadata" within that directory. If a list of paths, will assume that they make up a single parquet data set. Can also be any file-like object for a single-file dataset. |
| `verify`       | bool [False]                     | Test file start/end byte markers                                                                                                                                                                                                               |
| `open_with`    | function                         | With the signature `func(path, mode)`, returns a context which evaluated to a file open for reading. Defaults to the built-in `open`.                                                                                                          |
| `root`         | str                              | If passing a list of files, the top directory of the data-set may be ambiguous for partitioning where the upmost field has only one value. Use this to specify the dataset root directory.                                                     |
| `fs`           | fsspec-compatible filesystem     | You can use this instead of open_with (otherwise, it will be inferred)                                                                                                                                                                         |
| `pandas_nulls` | bool (True)                      | If True, columns that are int or bool in parquet, but have nulls, will become pandas nullable types (Uint, Int, boolean). If False, both kinds will be cast to float, and nulls will be NaN unless pandas metadata indicates otherwise.        |

### Attributes

| Attribute            | Type                | Description                                                                                                                                                                  |
| -------------------- | ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `fn`                 | path/URL            | Of '\_metadata' file                                                                                                                                                         |
| `basepath`           | path/URL            | Of directory containing files of parquet dataset                                                                                                                             |
| `cats`               | dict                | Columns derived from hive/drill directory information, with known values for each column                                                                                     |
| `categories`         | list                | Columns marked as categorical in the extra metadata (meaning the data must have come from pandas)                                                                            |
| `columns`            | list of str         | The data columns available                                                                                                                                                   |
| `count`              | int                 | Total number of rows                                                                                                                                                         |
| `dtypes`             | dict                | Expected output types for each column                                                                                                                                        |
| `file_scheme`        | str                 | 'simple': all row groups are within the same file; 'hive': all row groups are in other files; 'mixed': row groups in this file and others too; 'empty': no row groups at all |
| `info`               | dict                | Combination of some of the other attributes                                                                                                                                  |
| `key_value_metadata` | dict                | Additional information about this data's origin, e.g., pandas description, and custom metadata defined by user                                                               |
| `row_groups`         | list                | Thrift objects for each row group                                                                                                                                            |
| `schema`             | schema.SchemaHelper | Print this for a representation of the column structure                                                                                                                      |
| `statistics`         | dict                | Max/min/count of each column chunk                                                                                                                                           |

### Methods

#### `to_pandas()`

```python
ParquetFile.to_pandas(columns=None, categories=None, filters=[],
                       index=None, row_filter=False, dtypes=None)
```

Read data from parquet into a Pandas dataframe.

**Parameters:**

| Parameter    | Type                                       | Description                                                                                                                                                                                                                                                                                     |
| ------------ | ------------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `columns`    | list of names or None                      | Column to load (see `ParquetFile.columns`). Any columns in the data not in this list will be ignored. If None, read all columns.                                                                                                                                                                |
| `categories` | list, dict or None                         | If a column is encoded using dictionary encoding in every row-group and its name is also in this list, it will generate a Pandas Category-type column. If a dict {col: int}, the value indicates the number of categories. If None, will automatically set if the data was written from pandas. |
| `filters`    | list of list of tuples or list of tuples   | To filter out data. Filter syntax: `[[(column, op, val), ...],...]` where op is `[==, =, >, >=, <, <=, !=, in, not in]`. The innermost tuples are transposed into a set of filters applied through an AND operation. The outer list combines these sets of filters through an OR operation.     |
| `index`      | string or list of strings or False or None | Column(s) to assign to the (multi-)index. If None, index is inferred from the metadata.                                                                                                                                                                                                         |
| `row_filter` | bool or boolean ndarray                    | Whether filters are applied to whole row-groups (False, default) or row-wise (True, experimental).                                                                                                                                                                                              |
| `dtypes`     | dict                                       | Override dtypes for columns                                                                                                                                                                                                                                                                     |

**Returns:** Pandas data-frame

---

#### `iter_row_groups()`

```python
ParquetFile.iter_row_groups(filters=None, **kwargs)
```

Iterate a dataset by row-groups. If filters is given, omits row-groups that fail the filter (saving execution time).

**Returns:** Generator yielding one Pandas data-frame per row-group.

---

#### `count()`

```python
ParquetFile.count(filters=None, row_filter=False)
```

Total number of rows. Filters and row_filters have the same meaning as in to_pandas. Unless both are given, this method will not need to decode any data.

---

#### `head()`

```python
ParquetFile.head(nrows, **kwargs)
```

Get the first nrows of data. This will load the whole of the first valid row-group for the given columns.

**Parameters:**

- `nrows`: Number of rows to return
- `**kwargs`: Can include things like columns, filters, etc., with the same semantics as to_pandas()

**Returns:** dataframe

---

#### `write_row_groups()`

```python
ParquetFile.write_row_groups(data, row_group_offsets=None, sort_key=None,
                              sort_pnames=False, compression=None,
                              write_fmd=True, open_with=<built-in function open>,
                              mkdirs=None, stats='auto')
```

Write data as new row groups to disk, with optional sorting.

**Parameters:**

| Parameter           | Type                         | Description                                                                                                                                        |
| ------------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `data`              | pandas dataframe or iterable | Data to add to existing parquet dataset. Only columns are written to disk. Row index is not kept.                                                  |
| `row_group_offsets` | int or list of int           | If int, row-groups will be approximately this many rows. If a list, the explicit index values to start new row groups. If None, set to 50,000,000. |
| `sort_key`          | function                     | Sorting function used as key parameter for row_groups.sort() function                                                                              |
| `sort_pnames`       | bool                         | Align name of part files with position of the 1st row group they contain                                                                           |
| `compression`       | str or dict                  | Compression to apply to each column                                                                                                                |
| `write_fmd`         | bool                         | Write updated common metadata to disk                                                                                                              |
| `open_with`         | function                     | When called with f(path, mode), returns an open file-like object                                                                                   |
| `mkdirs`            | function                     | When called with a path/URL, creates any necessary directories                                                                                     |
| `stats`             | True/False/list of str       | Whether to calculate and write summary statistics                                                                                                  |

---

#### `remove_row_groups()`

```python
ParquetFile.remove_row_groups(rgs, sort_pnames=False, write_fmd=True,
                               open_with=<built-in function open>,
                               remove_with=None)
```

Remove list of row groups from disk. ParquetFile metadata are updated accordingly. This method cannot be applied if file scheme is simple.

---

## write Function

```python
fastparquet.write(filename, data, row_group_offsets=None, compression=None,
                  file_scheme='simple', open_with=<built-in function open>,
                  mkdirs=None, has_nulls=True, write_index=None, partition_on=[],
                  fixed_text=None, append=False, object_encoding='infer',
                  times='int64', custom_metadata=None, stats='auto')
```

Write pandas dataframe to filename with parquet format.

### Parameters

| Parameter           | Type                             | Description                                                                                                                                                   |
| ------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `filename`          | str or pathlib.Path              | Parquet collection to write to, either a single file (if file_scheme is simple) or a directory                                                                |
| `data`              | pandas dataframe                 | The table to write                                                                                                                                            |
| `row_group_offsets` | int or list of int               | If int, row-groups will be approximately this many rows. If None, set to 50,000,000                                                                           |
| `compression`       | str or dict                      | Compression to apply to each column, e.g. `GZIP` or `SNAPPY` or a dict like `{"col1": "SNAPPY", "col2": None}`                                                |
| `file_scheme`       | 'simple', 'hive', or 'drill'     | If simple: all goes in a single file. If hive or drill: each row group is in a separate file, and a separate file (called "\_metadata") contains the metadata |
| `open_with`         | function                         | When called with f(path, mode), returns an open file-like object                                                                                              |
| `mkdirs`            | function                         | When called with a path/URL, creates any necessary directories                                                                                                |
| `has_nulls`         | bool, 'infer' or list of strings | Whether columns can have nulls                                                                                                                                |
| `write_index`       | boolean                          | Whether or not to write the index to a separate column. By default we write the index if it is not 0, 1, ..., n                                               |
| `partition_on`      | string or list of string         | Column names passed to groupby in order to split data within each row-group, producing a structured directory tree                                            |
| `fixed_text`        | {column: int length} or None     | For bytes or str columns, values will be converted to fixed-length strings                                                                                    |
| `append`            | bool or 'overwrite'              | If False, construct data-set from scratch; if True, add new row-group(s) to existing data-set                                                                 |
| `object_encoding`   | str or {col: type}               | For object columns, gives the data type. Possible values: bytes, utf8, json, bson, bool, int, int32, decimal, infer                                           |
| `times`             | 'int64'                          | How to store time data                                                                                                                                        |
| `custom_metadata`   | dict                             | Custom metadata to add to the file                                                                                                                            |
| `stats`             | True/False/'auto'/list of str    | Whether to calculate and write summary statistics                                                                                                             |

### Compression Options

The compression parameter accepts:

- A string like `'GZIP'` or `'SNAPPY'` to apply to all columns
- A dict like `{"col1": "SNAPPY", "col2": None}` for per-column compression
- A dict with compressor arguments:

```python
{
    "col1": {
        "type": "LZ4",
        "args": {
            "mode": "high_compression",
            "compression": 9
        }
    },
    "_default": {
        "type": "GZIP",
        "args": None
    }
}
```

---

## update_file_custom_metadata Function

```python
fastparquet.update_file_custom_metadata(path, custom_metadata)
```

Update custom metadata in a parquet file.

**Parameters:**

- `path`: Path to the parquet file
- `custom_metadata`: Dictionary of custom metadata to add/update
