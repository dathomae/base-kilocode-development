# System Shape Discussion

The extended discussion on https://github.com/horatiar/chronocone/pull/19 has brought me to the
realization that we're not on the same page with respect to many of the basic features of ChronoCone,
our understanding of the potential users, the user experience and possibly the basic philosophy/idea
behind the system.  Obviously these conflicts need to be resolved.

A general note: I'm going to make some statements about my understanding of various things
and express some opinions in this document which are based on that understanding. Rather than 
litter this documents with "in my opinion" or "as I understand it" kinds of verbiage I'm just
going to say things. I'm explicitly NOT implying that this is the way things have to be, but I am trying
to maximize the probability that someone would read something in this doc and say "that's not correct"
because that will get closer to the root cause of any misalignment.

# Personal Philosophy
That said, I have a couple of viewpoints that I'm going to lean into more heavily:

1) All discussions should begin with the user/customer (not always the same people, despite
what Amazon thinks) experience and go from there. That experience mostly doesn't have a v0 vs. 
later verions type of dicussion associated with except possibly for small details. The questions in 
this doc are mostly of the customer experience type, at least to start with, although in some
cases they're informed by what's possible for implementation and will obviously lead to 
architecture/design/implementation discussions.
2) I don't much care if errors/poor phrasing, poor decisions or complexity that doesn't 
pay for itself and other issues come from the human user or are AI generated (I'm purposely not
using the word 'hallucination' here, that means something specific to me - many AI generated
mistakes aren't due to it making stuff up, but I digress), a mistake is a mistake. I 
take responsibility for the stuff I generate. When fleshing out the big ideas quickly like we're doing
now I'm going to concetrate on the main points and not worry about the little stuff as much (although
that will ultimately need to be brought into line, of course), but it's
always fair game to point out any problem and it's not necessary to distinguish between what I might
have personally generated vs. what AI generated. 
3) Before the age of AI the sheer labor involved in producing docs and code put a constraint on 
how much I'd be willing to throw way and start over. That constraint is much looser now, so I'm
MUCH more willing to throw bigger things away, since the labor required to create a revised form
is less WITH THE CONSTRAINT that as a default assumption the user experience should remain the same for the major aspects of the
system from version to version.  Adding new stuff on is fine, and there are bound to be some ideas
that don't work out and need to be changed, but the product should be recognizable as the same thing
from version to version.


## My Mental Model Of The Users

My mental model of the users has them falling into two main camps:
1) The ultimate end users who are running their organizations, who ultimately define
what the goals are and who ultimately pay for the software. My assumption is that
these users will be deeply knowlegable in the domain of their business and will sometimes
wind up creating or modifying models, possibly after a hand-off from a contractor/consultant.
These users will be primarily business analysts, with some programming/Python capabililties.
While they'll reach out to their consultants for big things, once a set of plans is in place
they'll want to make smaller changes themselves and will place a particular emphasis on ease of use,
the enforcement of guardrails and process to prevent things from getting screwed up which 
motivates their desired for asset versioning and lineage tracking both in a 'how was this data
value come from and how was it calcuated' sense and in a 'what changed by who and when' sense. They'll
also want easy identification of root causes and easy remediation (e.g., plan change rollback) if
when things get messed up. They pretty much want planning to just work, with minimal muss and fuss. In the
age of AI their primary interest in AI will be to ask questions and do analysis of the effects of plans
and in understanding plans. There will also be an interest in using AI to modify or create plans, but it will
be a less intensive or intricate activity than that of contractors.
2) Contractors/consultants who are essentially the intermediary between ultimate end users and
us. My assumption is that they will be deeply knowlegable in the general domain of whatever
business their customers are in, but less knowlegable about the details of any particular business.
They'll begin any engagement with a particular customer by gathering requirements, making proposals
and so on in the traditional way. I also assume that while they won't necessarily be software engineers
or developers they'll have a stronger background in programming/Python and so on. In practice contractors
operate in two distinct modes:
a) Initial engagement and "major change" mode. The contractor is heavily involved, their main concern is
convincing their customers that what they come up with works for whatever definition of "works" the customer
has. This work is often done under heavy time/resource pressure, so the contractors place a strong empahsis
on fast development and easy identification of issues and how to fix them. Once the initial engagement (or 
major re-engagement) is done, the contractor will mostly hand off control and maintainence of the system
to their customer at the customers insistent (i.e., they don't want to pay the contractor as an on-going routine cost).
Not all of the contractors customers will behave that way, but particularly the smaller ones that are likely
to be the major focus of early Chronocone deployments will.
b) "what the hell" engagements. If/when the ultimate customer has a problem that they can't figure out
they'll come back to the contractor, so these are problems that are harder to solve or require deeper insight 
into the system. They won't necessarily be handled by the same people who worked on the initial engagement or major change.
There's an emphasis in this mode of easy/quick understanding of what the system does, how it's organized, what the
goal is and the specific details of the ultimate customer that aren't implied by broader domain knowledge in the
business area. There's a strong emphasis on understanding and resolving problems quickly. A subset of these
problems will come back to us. This could be a prime application area for AI.

# Areas of Misunderstanding From The User/Customer Point Of View
Based mostly on the recent PR discussion these are areas where the conversation is causing me confusion or 
disagreement. Except for things where there are fundamental physical constraints on the operation of the system these
aren't about the exact implementation mechanism nor about v0 vs. later versions tradeoffs. These are questions
about how the system should behave from the end user viewpoint. I've put these in order of how fundamental they
are based on my current understanding. They're not entirely orthoganal, so it's possible that some of the later
questions may not make sense once the questions earlier in the list are resolved.

1) "Logic travels; data stays: on-prem calcuator".  This needs some clarification. I completely understand that 
potential customers are concerned about their data travelling to places they don't want it to go to the point of paranoia, 
but that may imply contraints that I'm not sure are consistent with our current ideas. Assume, to start with,
that data will encrypted both in transit and at rest, whereever it goes, and that we're managing key generation and storage in a secure way. 
Also assume that to be usable data will be decrypted and readable and that the decrypted form will reside in memory (e.g., inside a 
browser or plugin) or in the case of things that use an LLM sent to the LLM.

Is this implying that deployments will only happen on-prem (or to stretch a point, in an ultimate end users private cloud
environment) and that data from that environment will be accessible only in that environment? Is this specifically a statement
about the calculation of metrics or does it imply something about other data mechansims. There are three main touchpoints that
I can see:
a) Browser delivered data will generally go to browsers running on people's laptops and that some of these people will be
physically remote from any on-prem deployment. We can certainly push the calculations themselves back to the servers, but distinguishing
the results of the calcuation from the calculation itself would be splitting hairs from the users viewpoint. If there's an option 
that's not the 'in transit' encryption/decryption in the client approach than I'm not understanding it.  Also, does this means
that the browser would never support data download to .csv and things like that?
b) The idea of using an IDE plugin has the same questions, but but with the added complication that we can't control which plugins the user has 
installed and therefore it's possible that some data might wind up being logged or some such on the IDE users computer (technically
the user of browser extensions causes the same problem in broswer delivered apps, but IT orgs have tools and processes to deal with
that when they care, I don't think that's our problem to solve). There's a constraint that if code is to run locally (which is necessary
to support debug mode) with real data then the calcuation can't be pushed to a central server. There's nothing fundamental about Chronocone that 
says that we have to provide an IDE plugin, we could just create editing functionality in a browser based experience.
c) The AI/chat experience. This one is more fundamental because in any practical sense given the current state of the technology it requires that 
data be sent to an off-prem LLM. Either we don't expect to ever send data to an LLM (in which case I'm not sure that there's a reasonable case
for including any kind of agentic AI interaction in the product) or we need to acknowledge that data is going off-prem under some conditions and
making the product messaging consistent with that. Agentic AI use is inherently conversational, it's not a 'ask one question
and get one result'. I can't see that the mechanism suggested in the PR of asking a question and getting the result as a link is any kind
of solution - most any user would just expect for the AI to scrap the link anyway, in the end having an agentic conversation that just 
introduces a clunky data exhange mechanism in the middle. If some customer doesn't want any data to be sent to any LLM that's fine, we can
make that an option that needs to be enabled - but we should be clear that it's disabling the agentic capabilities, I strongly doubt that
we're going to be able to find any clever workaround that makes the AI experience compatible with user expectations that doesn't involve
multiple round trips to an underlying LLM without a human inspecting/approving each one. We should, of course, including config options
that let's the customer choose which LLM(s) to use, since some are going to promise not to retain data and others won't make that promise.
However, even then the message of "Logic travels; data stays: on-prem calcuator" should be adjusted to take into account customers who
want to use the agentic capabilities and are willing to send data off to carefully selected remote LLMs.

If the comment is specifically about the use of the MCP protocol rather than AI data exhange in general and it can't be addressed without 
compromising the experience expectation of reasonable users then we should just find a different mechanism.

2) "Executing a subset of Python".  I'm not sure how to interpret this or what motivates it, but specifying "which subset" would really
help the discussion. I can think of two possibilities:
a) It's motivated by the need to make that data doesn't escape by blocking certain calls, maybe print statements that reference data, things like that. 
It's essentially just another sandboxing mechanism, I understand that.
b) It literally means only allowing a subset of the language itself, rather than a set of calls - we're going to disallow the use of 
multi-level else statements or comprehensions or something (I'm having trouble thinking of examples because I don't understand the motivation).
That seems to me to be a bad idea, violating user expectations of what it means to write Python code by introducing (to the user) wierd exceptions. If
this is really what's intended then we might be better off using a DSL or something.

3) Plan consistency, your objection to the statement 'The DAG includes all operations needed to calculate every published metric in the namespace' (you should read the hypothesis about naming collision before continuing*). There's a thread around DAG calculation that seems to hang on the phrase "calcuate the whole DAG at once" that's causing 
confusion. Here's my perspective (which may be incorrect): From a user viewpoint a plan constitutes an interrelated set of metrics in some namespace family/namespace/"published/unpublished/dev sandbox" combination. In order for a plan to be internally consistent we should never show a partially updated plan to a
user (with the possible exception of some kind of debug mode(s), which I'm going to ignore here). Given that constraint we would always (again from the user
point of view) ALWAYS update an entire DAG in one of those subspaces at once (i.e., not update a single metric in the metric table resulting in a plan calcuation that's
internally inconsistent). I would not have considered that a controversial statement. 

*I suspect that the confusion is rooted in overloading the word 'namespace'. We have the Chronocone notion of namespace, but the combination of namespace family/namespace/"published/unpublished/dev sandbox" combination is a namespace in the software engineering sense. In the PR you highlighted the line 'all operations needed to calculate every published metric in the namespace', which is using namespace in the software engineering sense, but not the Chronocone sense. If this hypothesis is correct, we could try to handle the naming overload by careful wording of docs and tasks, but I think that we'd be better off coming up with new Chronocone terms for what is currently "namespace family" and namespace (doubly so because AI is almost guarenteed to interpret "namespace" in
the software engineering sense rather than the Chronocone sense). Maybe something like "Plan Type" for namespace family and "Deployment Type" or "Environment Type" for Chronocone namespace?

4) Metric lifecycle. I'm confused about the discussion of the Push/Publish/Update lifecycle for metrics. You made a comment about upsert_transformer/upsert_shifter, but those things seem to deal with datastore updates of those things, not the user visible lifecycle.

5) Relationship between metrics and transformers. I thought we'd agreed that transformer names would always be the same as metric names, which implied to me that there's a 1:1 relationships between metrics in a 'software engineering sense' namespace. Thus, defining a transformer isn't different from a users point of view than defining a metric (plus or minus stuff around the edges like defining shifters and actuals vs. forecast data pulls ala the demand forecast example where the forecast - not calculated by Chronocone - and the actuals don't come from the same source). 

# Internal Issues
Some of these are v0 vs. later discussions...

1) PR grain and work coordination. I've been working through the open question doc, but in hindsight it would have been better to do a PR after every increment with stories (or potentially some other mechanism) to go with them. It keeps the reviews more conceptually focused and so on. However, if your fundamental objection is that it 
takes too much time to review them then that won't help with that problem (the only idea that I had in this regard was breaking the docs into a hierarchy of detail, where we walk the tree downward in terms of reviews and discussion). We also need to agree on a coordination mechanism (I see you using stories in the stories directory, so maybe that's the answer), particular since using the AI approach 
there'll be multiple stories/tasks running in parallel and we want to make sure we don't collide or introduce something that's inconsistent. 

2) We should talk through the 'call exec' vs. 'assemble code' approaches for calcuation. I see the second as strictly better than the first (i.e., both about the same time to implement, with better performance for the second), therefore there's no point in doing the first even as a short term expedient.

3) I think that we're having mixed discussions about the v0 implementation of calculations and the eventual (more efficient, more scalable) version. The docs I wrote were most oriented to the v0 algorithm intended to run locally and implemetable in a short time version. I'm happy to talk about the longer term versions, but we should clearly distinguish between the two (if I'm wrong about that point, that's another thing to talk about).

4) Metric metadata storage in the metric. The intent is to adopt a simple mechanism to ensure that plans can always be interpreted, even if the definitions of metadata elements change and aren't consistent between the various 'software engineering sense' namespaces. We can do that through metadata versioning too, I just wasn't sure I was seeing enough of a mechanism in the db schemas to implement that, particuarly when definitions differ between 'software engineering sense' namespaces - but possibly I'm missing something. I agree that the metric definitions should always appear in one place, we shouldn't introduce the problem of reconciling disparate definitions that don't come from the user.























