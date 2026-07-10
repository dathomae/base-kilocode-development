A prompt driven development or PDD script is a pseudo-executable script intended for "execution" by an agentic tool. It is analogous to a shell script in the pre-LLM technology world.

A PDD script contains the following sections:
1) A description of the task or operation to be performed.

2) A list of parameters are required to execute the script.  Parameters may be required or optional, although all optional parameters MUST have a default value specified in the script. Parameters should be defined as a table with the following columns: Name, Type, Description, Required (Yes/No), Default Value (if optional).

3) A list of operations that MUST be performed sequentially. The PDD script should stress that constraint. Operations should be numbered sequentially and use imperative language (e.g., "Read the file", "Update the code").

4) A specification of the output items and output format for a script. Some scripts will be execute for their side effects, so it is permissable to have a script where the output consists only of reporting whether the operation succeeded or not, however most scripts will output some information about the task (e.g., number of file updated).

5) Error Handling: If required parameters aren't specified, the user should be prompted. The operations in the script WILL NOT BE executed until/unless all required parameters have been specified. During execution, if an error is encountered, the execution of the script will be paused and the user will be asked if script execution should continue or if the script should be terminated.

See create-new-pdd-script.pdd.script.md for an example implementation.
