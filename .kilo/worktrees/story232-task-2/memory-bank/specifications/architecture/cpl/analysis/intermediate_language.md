## Calculation/Operations Framework In Chronocone

Generating metrics in Chronocone consists of four phases:

1. During the description assembly phase a complete Chronocone description is assembled from a combination of CPL files (which will be parsed and semantically analyzed) possibly including elements read read from some datastore (e.g., built-in concept definitions). At this stage the metrics and lookups described may produce multiple versions of the time series that they define, potentially differing in their dimensions, grain, length and data type.
2. The result of the description assembly phase is combined with a specification that specifies the desired output, including the following elements:
   - The set of metrics to be calculated.
   - The set of desired dimensions in that set of metrics which can't be deduced from the other metrics in that set given the set of definitions that have been obtained.
   - The desired interval specification for the data, including interval constraint patterns and associated duration.
   - The desired length of the actuals and/or forecast aspect(s) of the specified metric(s).
   - The desired type of each metric.
3. Once the specification is known a well defined set of metric definitions will be determined, possibly including the application of aggregators, spreaders, unit conversions and so on. The result of this step is captured in the form of a Chronocone Intermediate Language (CIL) which is a machine and language agnostic representation of the calculations that must be performed. In addition to the representation of the calculation the CIL will include some metadata, such as input dependencies and the aformentioned output specifiction.
4. The CIL is then used to generate an executable form of the metric calculation by one more more synthesizers. The inputs to the synthesized program are Parquet files (plus queries from the database?). The output is a Parquet file that includes the values of the specified output metrics, plus the intermediate metrics required to calculate the output metrics plus the data in the input files. In addition metadata is produced that is used in lineage tracking.

A key constraint is that during the description assembly phase and the synthesis phase the concept names and relationships may be used, but the concept values cannot assumed. The concept values, and optimizations that depend on them, must be determined when the synthesized executable is run.

This document describes the CIL.

Discussion: The synthesized form will need the values from concepts not capture in the CPL description being parsed. Does that come direct from the database or is that also a parquet input? We'll have to figure out what the metadata that supports lineage looks like. Probably that mostly is a record of the concepts structure and values at the time that the output is generated (or ids and time used by the validity ranges). The actual DAG is captured in the CIL structure itself.

## Output Specification Interpretation

If any of the elements of the output specification in #2 is missing or is underspecified (e.g., the output specification says that the inventory_actuals metric is desired with for periods that start at midnight on Mondays with a duration for each period of 1 week) then the most aggregated version of the metric will be generated given the contents of the description. For example, if an inventory_actuals metric is desired and a Product dimension is specified for the output and the source data obtain from a connector has a grain of [Country, Product, Warehouse] of type Count for intervals that are each a day long starting at midnight, and there is an aggregator included in the description for the transformation from [Product, Warehouse] to [Product] then the most aggregated version of the output metric is at the [Country, Product] level and that is what will be produced (i.e., data by country for each product). If there was also an aggregator that transformed from [Country, Product] to [Product] that would be used the the output would just be by Product. If the output grain specified cannot be deduced from the definition and input data (e.g., because there's an aggregator or spreader missing), the user will be informed of the dimensional transformations that would be required to produce the desired output that are missing.

Discussion: We could make the user give a complete specification of the output dimensions, grain and so on, but it seems like a rule like the above will be more user friendly.

## Definitions

The analysis phase requires knowledge of the concept relationships (but not the concept values), the units and type definitions. When the synthesized program is run the concept values are read, which allows the synthesized program to determine the representation of the data (e.g., order of dimensions for a metric) that can be processed most efficiently. Base type and the constraints that associated with any given value are associated with each operation so that the synthesized program can determine the most efficient representation of the data. Units information, including unit conversions, is need for the analysis, but after that has no influence on the synthesized program. However, units are also carried through for each concept and operation to store with the final outputs to facilitate lineage tracking, explainability of the result and help with debugging.

## Tensor Operations

The fundamental represenation of data in the CIL is a tensor, which is just a higher dimensional vector or matrix. When a tensor has zero dimensions it's a constant, when it has one dimension it's a vector, when it has two dimensions it's a matrix (all of which are tensors), and when it has more than two dimensions there's no special name for it, so it's just a tensor.

A tensor can be thought of as a one dimensional vector of the elements of the dimension below it. Thus, if we had a metric that has dimensions of Country and Product it will automatically have a compulsory dimension of time as well (i.e., data values that are indexed by actual and/or forecast time periods). As will be seen in the discussion of aggregators and spreaders there is considerable benefit of not thinking of time as a special dimension. The dimensions of a tensor will have an order for any given set of dimensions. In the example metric if the order is Product->Time->Country and Country has a set of values c1, c2 and c3 for any given time period, time has a set of values a1, a0, f0 and f1 for two actuals and two forecast periods and the notation c_0_a0 indicates the metric value for the a0 time period for country 0, then the lowest two levels of the tensor would be:

```
[
    [ c_0_a0, c_1_a0, c_2_a0]^T
    [ c_0_a1, c_1_a1, c_2_a1]^T
    [ c_0_f0, c_1_f0, c_2_f0]^T
    [ c_0_f1, c_1_f1, c_2_f1]^T
]
```

Note: The bottom level vectors are actually column vectors (not row vectors as is shown here), but that would take a log of vertical space to write in this form, so I'm writing them here as transpose of row vectors. At some point we should rewrite this section using Latex formulas.

The first lowest level vector has three data values for country all of which correspond to a single value of the dimension above it, which in this case is time in the form of the forecast periods for the metric time series.

If the product dimension has two possible values x and y and the value for c0 at time a0 for product px was c_0_a0_x then the whole tensor would be:

```
[
    [
        [ c_0_a0_x, c_1_a0_x, c_2_a0_x]^T
        [ c_0_a1_x, c_1_a1_x, c_2_a1_x]^T
        [ c_0_f0_x, c_1_f0_x, c_2_f0_x]^T
        [ c_0_f1_x, c_1_f1_x, c_2_f1_x]^T
    ]
    [
        [ c_0_a0_y, c_1_a0_y, c_2_a0_y]^T
        [ c_0_a1_y, c_1_a1_y, c_2_a1_y]^T
        [ c_0_f0_y, c_1_f0_y, c_2_f0_y]^T
        [ c_0_f1_y, c_1_f1_y, c_2_f1_y]^T
    ]
]
```

Because Product is at the top level of the tensor representation in this case the number of the top two vectors will be the same as the number of values associated with Product (2 in this case), each of those will have a length that is associated with the next level dimension of time (4 in this case) and the bottom level vectors have a length which is the same as the bottom level Country dimension (3 in this case). Thus the tensor has the shape of (2x4x3).

When operating on two tensors, to be valid the two tensors involved must always have dimensions size above the bottom two levels that are the same. Thus, if the order of the dimensions was Country->Product->Time the tensor would have a shape of (3x2x4), and because there's a dimesion above the bottom two that isn't the same we know that there are no operations that could be valid for it. We can, however, _pivot_ any tensor to reorder it's dimension in order to reshape it.

The bottom two vectors taken together are a 2D matrix, and they can be reshaped using matric operation rules and that is the reason for adopting this tensor formulation. For example, if one wanted to remove the country dimension by adding up the three country values for each of the time dimensions, the matrix operation:

```
[1 1 1][<the tensor above>]
```

would produce exactly that result. If the multiplying matrix was placed to the right of the tensor, then it would expand the lowest level dimension into something else, and thus add a dimension to the tensor. This will be discussed in more detail in the section on aggregators and spreaders.

While the use of direct tensor operations in the generated executable form is encouraged if available, when an executable form (e.g., a Go program) is created that does not directly support vector operations it is permissible to combine more primitive operations to implement the vector operations (e.g., by generating loops or calling utility functions). The intent is that the CIL will supply information that will allow the generated form to create highly optimzed code, while still supporting a variety of executable forms.

## CIL Concepts

The Chronocone Intermediate Language (CIL) consists of a set of trees of operations, where each tree is rooted in one of the 'output' metrics. An output metric is a metric that is included in the calculation only because it appears in the output specification (it might also be referenced by some other output metric, in which case it's inclusion in the output specification is redundant). All other metrics in the CIL trees for a particular output specification are present because they're required to calcuation the output, either directly or indirectly. While the original description of a set of metrics may be a cyclic graph at the metric level and a DAG at the time series cell level the analysis that creates the intermediate represenation will choose an order of calculations to make them into a pure tree representation.

Another important concept is that the time dimension of a time series is treated as a compulsory dimension that is treated in the same way as any other dimension in all other ways. Like every other dimension the dimension values have a certain type, which implies certain operations that may be performed on the dimension values, and a collection of values. The motivation for this perspective is that it allows the executable that generates a document the latitude it needs to organize it's data in a way that is most efficient for the document generation process. For example, imagine that we have a metric that depends on Country, Product and the built-in concept of the sequenced time dimension that all metrics have. Assume that Product has 1000 values, that Country has 3 values and that the metric has a time grain of a week and has 13 weeks of actuals and 13 weeks of forecast. To compute results efficiently, the executable form would like to assemble groups of numbers into contiguous memory blocks of a size that will permit maximum cache utilization/locality. If we built the idea of time as a special dimension into the generated code as a 'metric time major form' then for the combination of the Product values and the time series in the Time dimension the calcuation would do 1000 calculations on a block of 26 numbers each (essentially iterating through the time series cells in time order), which is likely to be much less than the cache size and thus lead to more frequent cache misses. However, if that same calculation is organized in 'Product major form' then that same calculation would be done on 26 blocks of 1000 numbers each (effectively doing the calculation for all of the Product values for a given period in the metric). However, using that same scheme, the data for the metric could be represented as a single block of 26000 numbers _that can all be operated on in the same vector operation_ since the sections of the data that represent the same time periods for the same metrics would all line up. The generated form can actually do better than that, but the details of that discussion will be left to the synthesation documents.

While the CIL uses a functional formulation, it is not intended that the synthesized code will actually be in a functional form as long as the results of the synthesized calculation are identical to the result of computing the functional form are identical. In fact, it's anticipcated that the synthesized form will frequently read inputs to a calculation from a memory location, perform the operation and write the result back into the same memory location in order to facilitate the goal of doing as many reads and writes as possible to cached memory locations.

Imagine that we're performing these calculations on a computer that uses an i7-14700K processor, which has 20 physical cores and an L2 cache for each core that has a length of 33MB (actually that's the length for the P cores, but ignore that for the purposes of this example). There are 26000 numbers for each of the two inputs and the output. If each of the numbers is a 64 bit integer, then each of the inputs and the output is ~203KB in size. If we divide them each into 21 chunks and send each chunk to a different core, there will be room for each of the inputs and the output in the cache. For a single operation most of time might be taken up loading the data into the L2 cache. However, when multiple operations can be chained together (without changing the dimensions) where the result of each calculation is the input to the next calculation, the most time consuming part of computation, namely shuffling the data around between cache and main memory, can be avoided. Because different metrics along a calcuation chain will likely bring new metrics in along the way it's probable that we'd want to divide the available cache into a number of larger chunks so that the data for the different metrics in a calculation chain can be loaded into the cache at the same time followed by a series of calculation that just reads and writes to the cache (this is essentially a register allocation problem). While this document won't treat the details of data structures and operations prediction that would make that kind of optimization possible, the constructs and structure of the CIL have been chosen to make that transformation easy. Similar concepts apply when doing metric calculations for a set of metrics where all of the data can't be fit into memory at once, requiring the use of streamed or paged IO during the course of the calculation.

In the actual CIL there will be a represention for the time dimension that can't be confused with any user concept definition. For the purposes of this example, we'll use the dimension name 'Time' for that concept for purposed of explanation, but it's not literally a concept definition.

### CIL Tensor Representation

Since the actual dimension and data values aren't available at analysis or synthesis time, the CIL doesn't assume any particular set of values, but it does assume that each data element has two parts:

1. Tensor data, which conceptually has one or more dimensions (since time is a compulstory dimension) where each combination of values for all of the dimensions is a single data value. While it will be functionally valid for the executable program to represent data in that form, the CIL idea of the structure of the tensor data should be regarded as a formalism. The intent is that the tensor data will be represented as a single dimensioned array of values, where the order of the dimensions is chosen at execution time by the synthesized program depending on the relative length of the dimension values.
2. A mask, which is a tensor that has boolean values with the same dimensions as the tensor data. The values in the tensor that are used will have a value of True in the mask, and the values that are not used will have a value of False. It is permissible for the synthesized program to set or ignore values in the tensor data that correspond to a False value in the mask in whatever way is convenient or performance enhancing in the synthesized program since by the time that any particular tree in the CIL has been traversed any masked out values will have been combined with other masked out values to produce a full tensor as the final result.

The combination of these two boolean mask tensor and data tensor elements will be referred to as a 'masked tensor'.

### CIL Fundamental Data Operations

The CIL supports the following fundamental data operations (meaning that they affect the values of data but not their shape):

1. map, which takes as input the name of a list of lists of concept values, a function, and a set of inputs which must be sets of masked tensors. map applies the function to each element of each unmasked value in each masked tensor that corresponds to each of the combinations of the values for the concept names and generates a single output which is a set of masked tensors that has the values produced by the function for each unmasked value. The output for masked values is undefined.
2. combine, which will take a pair of masked tensors as input assemble the two masked tensors into a single masked tensor, using the mask specifications to determine which values to use. It is an error for the same value in the output to have a value of True in both masks.

For example, suppose that we have a metric that has inputs from two other metrics, and at the point of the calculation the two inputs and the outputs share the same dimensions, time grain and units (as they must, because either the a sequence of zero or more aggregators, spreaders and unit conversions have been applied to bring that situation about or an error was produced telling the user that the output they requested can't be generated from the available description and output specification).

The general form of the map operation is:

```
(map [<values for Time>, <values for Country>, <values for Product>], <value generation function>, <mask generation function>)
```

where the `<mask generation function>` is an optional function that, if present, must return a boolean value for each combination of the values of the dimensions. If the mask generation function is not present, then the mask will be True for all values.

If each element of the two input metrics is added together to produce the corresponding element of the output metric, then the CIL for that calculation would be:

```
(map [<values for Time>, <values for Country>, <values for Product>], (add, (input1, input2)))
```

In order to avoid cluttering the examples, in the rest of this document we'll write this as

```
(map,
 [Time, Country, Product],
 (add,
   (input1, input2)))
```

Since the combination of of the values of the dimensions must always result in a single value, the result of each individual operation will also have as input a single value of each of the inputs, producing a single value of the output. The map operation finds the appropriate input values for each combination of the values of the dimensions, applies the given operation and assembles the results into a masked tensor. With the understanding that the CIL representation of a constant is as a zero dimensional tensor (i.e., all dimensions have a length of 0), all of the inputs must be masked tensors and the output is a masked tensor. It's anticipated that the executable form generated from the CIL plus input specification will represents tensors as either multiple dimensional arrays or a single array and use arithmetic expressions to index into the array for different dimension values (including the compulsory time dimension). It is free to choose any order of dimension values that produces valid results, but it's anticipated that it will group time series cell values so that the values that correspond to the largest set of concept values are stored contiguously in memory.

Various types of low level operations are supported for data in a time series cell and for dimensional values.

### Supported Data Operations Functions

Note: The only operation types that are supported for cell data operate on Numbers and Strings. If we need to work with other types of data such as Intervals (as data in a time series cell, not dimension values) or Booleans, we'll need to add additional operators to support that/those type(s), but the overall form of the CIL representation will remain the same.

| Operation | Operation Type | Input Type(s) | Cell Values/Dimensional Values | Description                                               |
| --------- | -------------- | ------------- | ------------------------------ | --------------------------------------------------------- |
| negate    | arithmetic     | Number        | Cell                           | Negates the input. May only be applied to a single input. |
| add       | arithmetic     | Number        | Cell                           | Numeric sum.                                              |
| add       | arithmetic     | String        | Cell                           | Concatenates the inputs.                                  |
| subtract  | arithmetic     | Number        | Cell                           | Numeric difference. May only be applied to two inputs.    |
| multiply  | arithmetic     | Number        | Cell                           | Numeric product.                                          |
| divide    | arithmetic     | Number        | Cell                           | Numeric quotient. May only be applied to two inputs.      |

### Boolean Operations

Consider a metric with the following definition that includes boolean operation in it's definition:

```
metric make_it_even of Count
  inputs
    inventory
  calculate
    if inventory % 1 = 0
      inventory
    else
      inventory + 1
    end if
```

In an imperative language we'd think of the calculation expression in terms of program flow. The CPU at some point would first execute the instruction that calculates the mod value for a number. Then a boolean operation would be applied and program flow would either continue on to read the input value and brach to the instruction following the if statement or continue on to read the input value, add one to it and continue on to the next instruction after the if statment. However, because the CIL considers all calculations to be tensor expressions that potentially may be performed in parallel at any given level of an expression tree, a conditional expression with one if and an else splits the inputs into two sets, the set to which 1 will be added and the set to which it will not be. The boolean operations set the values in the mask portion of a masked tensor, while the arithmetic operations supply the data values in the data tensor portion of the masked tensor.

If equal_to is a boolean operation that causes a value in the mask portion of a masked tensor to be set to true when it's arguments are equal, then we can calculate the values in the first clause in a tensor with mask values set to True only for those data elements were the condition evaluates to true, calculate the values in the second clause in a tensor with mask values set to True only for those data elements were the condition evaluates to false, and then combine the two tensors.

This operation would be represented in the CIL as follows:

```
(combine
    (map
        [<list of dimensions for the metric],
        (add inventory 0),
        (equal_to ((mod inventory 1), 0))
    ),
    (map
        [<list of dimensions for the metric],
        (add inventory 1),
        (not (equal_to ((mod inventory 1), 0)))
    )
)
```

The original if statement has two branches, so there are two subexpressions passed to the combine operation. The first subexpression is the expression that will be evaluated if the condition is true. The second subexpression is the expression that will be evaluated if the condition is false. Note that the boolean expression for each subexpression must stand alone. In this case a simple not of the first expression is all that is required, however those boolean expressions will become more complex if there were multiple `else if` statements or multiple clauses in a `when` statement. When values for the first subexpression are calculated, boolean mask operation is conceptually executed first and the result used to set the values in the mask part of the masked tensor and then the values for at least the True values in the mask tensor are generated and written into the data tensor part of the masked tensor. Note that the data values for the False values in the mask tensor are undefined. The synthesized program can generate them or not, or even use a sparse represenation of the tensors or not as it sees fit to maximize performance. The fact that the map operations in the combine have independent boolean operations means that the different inputs to the combine operation can be calculated in parallel, and that property holds true for every level of the expression tree.

--- Begin Discussion ---
The decision to require that the boolean operations for the inputs to a combine embodies an assumption about the relative time cost of the boolean operations vs. the time cost of shuffling data in and out of memory, which is that moving data is more expensive than boolean operations. It also facilitates sending chunks of the operation trees to different hosts should we want to do that (although larger chunks would need to be sent to keep the communication times over networks from overwhelming the benefit of parallelization). The consequence is that the boolean operations themselves would get more complex as more branches are involved. In the example there are only two brances, so the second boolean expression is just a not of the first. However, if we had the following:

```
  when
    input1 > input2 => <do whatever>
    input1 > input3 => <do whatever>
    default => <do whatever>
  end when
```

which would be equivalent to:

```
if input1 > input2
    <do whatever>
else if input1 > input3
    <do whatever>
else
    <do whatever>
end if
```

would result in the CIL expression:

```
(combine
    (map
        [<list of dimensions for the metric],
        (<the operation in the first clause>),
        (greater_than input1 input2)
    ),
    (map
        [<list of dimensions for the metric],
        (<the operation in the second clause>),
        (and (greater_than input1 input3) (not (greater_than input1 input2)))
    ),
    (map
        [<list of dimensions for the metric],
        (<the operation in the second clause>),
        (not (and (greater_than input1 input3) (not (greater_than input1 input2))))
    )
)
```

where each subsequent boolean expression captures the inverse of the earlier operations and will therefore tend to become more complex, although there is also a greater likelyhood of being able to simplify the expression.
Since it's an error for the same masked elements in any set of expression passed to a combine operation to have multiple inputs have True at any given position or to have any given position not have some mask values that is True the result is that the output of a combine will always be a full tensor.

--- End Discussion ---

The following boolean operations are supported for the mask function in map calls:

| Operation             | Operation Type | Input Type(s)  | Cell Values/Dimensional Values | Description                                                                         |
| --------------------- | -------------- | -------------- | ------------------------------ | ----------------------------------------------------------------------------------- |
| not                   | boolean        | Boolean        | Cell                           | True if the input is False, False otherwise.                                        |
| and                   | boolean        | Boolean        | Cell                           | True if both inputs are True, False otherwise.                                      |
| or                    | boolean        | Boolean        | Cell                           | True if at least one of the inputs are True, False otherwise.                       |
| equal_to              | boolean        | Number, String | Cell                           | True if both inputs have the same value, False otherwise.                           |
| greater_than          | boolean        | Number, String | Cell                           | True if the first input is greater than the second input, False otherwise.          |
| less_than             | boolean        | Number, String | Cell                           | True if the first input is less than the second input, False otherwise.             |
| greater_than_or_equal | boolean        | Number, String | Cell                           | True if the first input is greater than equal to the second input, False otherwise. |
| less_than_or_equal    | boolean        | Number, String | Cell                           | True if the first input is less than the second input, False otherwise.             |

Discussion: We don't necessarily have to have operations like 'greater_than_or_equal' since those can be composed of operations earlier in the list. However, they will be primitive operations in pretty much any
language we synthesize to and it would be a bit of work for the synthesizer to collapse those trees to find the primitive operations. There is no inverted version of those functions (e.g., not_equal_to, not_less_than)
because recognizing an inversion (which is basically just an operation wrapped in a not) is trivially easy to recognize and it reduces the proliferation of functions.

### Dimension Shifting Operations

The previous sections dealt with operations that calculate the values of a tensor of a certain shape taking inputs that must be of the same shape. Chronocone also supports operations that change the shape of the tensor, and in all cases will wind up replacing some set of values in a tensor with other values (e.g., replacing the values in a tensor that correspond to various Country values with some aggregate of those values when a Country dimension is being remove from the tensor, adding a series of values calculated from a single value in the tensor multiple by a vector of weights when a Country dimension is added to a tensor). There may also be situations where dimension values are dropped, dropping the data and mask values from a tensor associated with those values.

The dimensional operations are different from the data operations in that they can affect both the shape and the values of data. The fundamental dimensional operations are:

1. matrix_operation takes an ordered list of the expected output dimensions, a function and two tensors and returns the product of the bottom two levels of the tensor using the usual matrix multiplication rules with the operator being used to combine the elmentwise operations. If the operator is sum, then the result would be the traditional matrix multiplication. To be a valid operation the output dimensions must differ only in least significant dimesion from the output dimensions, therefore there may be a pivot operation performed just before a matrix_operation to satisfy that constraint.
2. pivot takes as input a masked tensor and an ordered set of dimensions. It reorganizes the tensor so that it has the shape that corresponds to the number of values of each dimension in order given to prepare for a matrix_operation to follow. Note that a pivot that swaps the bottom two dimensions of a tensor is equivalent to a transpose operation.
3. shift takes as input a masked tensor, a shift direction of UP or DOWN (considering the leaf level dimensions of the tensor as column vectors) and a number of shifts. The length of the leaf level vectors will be the same as the length of the starting leaf level vectors, but the values in the vector will be displaced in the direction by the number of entries given by the number of shifts. A number of values equal to number of shifts at the end of the vector in the shift direction will be dropped, a number of values equal to the number of shifts of the 'zero' value for the type will be added at the end of the vector opposite to the shift direction.

When a dimesion is removed from the input dimensions it must always appear as the lowest level dimension in the input dimension specification. When a dimension is added to the output dimensions it must always appear as the lowest level dimension in the output dimension specification. In all cases the shape of the result must reflect the output dimension specification both in order and in the number of values at each level.

A 'pivot' takes as input a masked tensor and an ordered set of dimensions. It reorganizes the tensor so that it has the shape that corresponds to the number of values of each dimension in order given to prepare for a matrix_operation to follow. Note that a pivot that swaps the bottom two dimensions of a tensor is equivalent to a transpose operation.

For example, imagine that we have the [Product, Time, Country] tensor from the Tensor Operations section:

```
A = [
    [
        [ c_0_a0_x, c_1_a0_x, c_2_a0_x]^T
        [ c_0_a1_x, c_1_a1_x, c_2_a1_x]^T
        [ c_0_f0_x, c_1_f0_x, c_2_f0_x]^T
        [ c_0_f1_x, c_1_f1_x, c_2_f1_x]^T
    ]
    [
        [ c_0_a0_y, c_1_a0_y, c_2_a0_y]^T
        [ c_0_a1_y, c_1_a1_y, c_2_a1_y]^T
        [ c_0_f0_y, c_1_f0_y, c_2_f0_y]^T
        [ c_0_f1_y, c_1_f1_y, c_2_f1_y]^T
    ]
]
```

and an aggregator is defined that collapses the dimensions to [Product, Time] by aggregating the values for the countries for each time period. Assuming the input dimension order of [Product, Time, Country] That equates to the multiplication operation:

```
matrix_operation([Product, Time], sum, [1 1 1], A)
```

For the first lowest level vector the result is [1 1 1][c_0_a0_x, c_1_a0_x, c_2_a0_x] = [c_0_a0_x + c_1_a0_x + c_2_a0_x]. As a (1x1x1) tensor, that's equivalent to a constant, yielding the result:

```
B = [
    [
        [
          1*c_0_a0_x + 1*c_1_a0_x + 1*c_2_a0_x
          1*c_0_a1_x + 1*c_1_a1_x + 1*c_2_a1_x
          1*c_0_f0_x + 1*c_1_f0_x + 1*c_2_f0_x
          1*c_0_f1_x + 1*c_1_f1_x + 1*c_2_f1_x
        ]
    ]
    [
        [
          1*c_0_a0_y + 1*c_1_a0_y + 1*c_2_a0_y
          1*c_0_a1_y + 1*c_1_a1_y + 1*c_2_a1_y
          1*c_0_f0_y + 1*c_1_f0_y + 1*c_2_f0_y
          1*c_0_f1_y + 1*c_1_f1_y + 1*c_2_f1_y
        ]
    ]
]
```

which is what we want. If the desired result was a weighted average the form of the operation would be the same, but the first tensor would have values that were fractions that added up to 1 rather than [1 1 1]. Combining operations other than sum would give a type of quotient algebra, but I'm not sure that that would be useful in a business context. Therefore, while the possibiliy of defining other combining operations is avaiable, for the time being I'm only going to include sum.

Suppose that wanted to undo that operation using a spreader? The same matrix_operation operator would be used, again with the sum function, but with the multiplying matrix on the right hand side instead of the left hand side (assuming input dimension order of [Product, Time]):

```
matrix_operation([Product, Time, Country], sum, A, [1/3 1/3 1/3])
```

which would result in each of the leaf level numbers creating a vector of three values, each of which is the original value for each time period (obtained by the sum of the original three values) divided by three. When the country values were originally added information was lost, so that won't reproduce the original result, but in situations like this it's likely that some weighted division would be used, which would simply change the values in the constant vector to be something other than one third.

In the previous example an implicit assumption was made that the expansion from [Product, Time] to [Product, Time, Country] would use the same propotions for each Country/Product combination for each time period. What if the proportions were different for each Product value in each Country? In that case, the vector operation would be an operation between Product and Country, so Product would need to be the lowest level dimension on the input vector, requiring a pivot, assuming the input tensor had a dimension order of [Product, Time] and Country would need to be the lowest level dimension on the output vector.

```
matrix_operation([Time, Product, Country],
                 sum,
                 (pivot [Time, Product] A),
                 [[.4 .6]^T [.3 .7]^T [.1 .9]^T ])
```

To execute the operation the input would first be pivoted from [Product, Time] to [Time, Product] and then matrix multiplication will result in two values being created for each of the three country values.

The following table summarizes the functions that are supported byt the matrix_operation operator.

| Operation | Operation Type | Input Type(s) | Cell Values/Dimensional Values | Description                                                              |
| --------- | -------------- | ------------- | ------------------------------ | ------------------------------------------------------------------------ |
| sum       | arithmetic     | Number        | Cell                           | Adds the values of the individual vector/matrix entry products together. |
