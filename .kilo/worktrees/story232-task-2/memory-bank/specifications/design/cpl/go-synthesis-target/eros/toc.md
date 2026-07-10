# EROs - Go Synthesis Target

This directory contains Epic Restraining Orders (EROs) for the Go synthesis target of the Chronocone Planning Language (CPL). EROs capture design decisions specifically related to managing confusion from incremental implementation.

## EROs

**[ERO001_concepts-synthesis-phase.md](ERO001_concepts-synthesis-phase.md)** — _Draft_
: Documents how concepts are handled in the synthesis phase

**[ERO002_types-synthesis-phase.md](ERO002_types-synthesis-phase.md)** — _Draft_
: Documents that types require no synthesis phase work

**[ERO003_units-measures-synthesis-phase.md](ERO003_units-measures-synthesis-phase.md)** — _Draft_
: Documents that units and measures require no synthesis work

**[ERO004_connectors-synthesis-phase.md](ERO004_connectors-synthesis-phase.md)** — _Draft_
: Documents how connectors are synthesized to read parquet data

**[ERO005_metrics-synthesis-phase.md](ERO005_metrics-synthesis-phase.md)** — _Draft_
: Documents how metric CIL trees are walked to generate Go code

**[ERO006_lookups-synthesis-phase.md](ERO006_lookups-synthesis-phase.md)** — _Draft_
: Documents how lookup CIL nodes are walked to generate Go code for data loading and runtime access

**[ERO007_aggregators-synthesis-phase.md](ERO007_aggregators-synthesis-phase.md)** — _Draft_
: Documents tensor matrix multiplication, aggregator CIL node walking, and metric integration

**[ERO008_spreaders-synthesis-phase.md](ERO008_spreaders-synthesis-phase.md)** — _Draft_
: Documents spreader CIL node walking, dimension expansion, and metric reference computation
