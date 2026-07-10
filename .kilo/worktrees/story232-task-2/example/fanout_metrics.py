from typing import Annotated
from chronocone import In, Out, Metric

@Metric(ns_family="sales", ns="shock forecast", transform="funnel conversion")
def arbitary_function_name(
        funnel_top: Annotated[float, In(metric="funnel_top")],
        funnel_middle: Annotated[float, In(metric="conversion_rate")]
) -> Annotated[float, Out(metric="simplified customer expectation")]:
    if funnel_top < 0.0:
        funnel_top = 0.0
    
    funnel_middle = max(0.0, funnel_middle)
    return funnel_top * funnel_middle
