# test-network-function Catalog
The catalog for test-network-function contains a variety of `Test Cases`, as well as `Test Case Building Blocks`.
 * Test Cases:  Traditional JUnit testcases, which are specified internally using `Ginkgo.It`.  Test cases often utilize several Test Case Building Blocks.
 * Test Case Building Blocks:  Self-contained building blocks, which perform a small task in the context of `oc`, `ssh`, `shell`, or some other `Expecter`.

So, a Test Case could be composed by one or many Test Case Building Blocks. 

