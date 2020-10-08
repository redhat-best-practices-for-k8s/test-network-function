# Getting Started with CNF Certification

1. Download and install up-to-date oc client:
   <https://access.redhat.com/documentation/en-us/openshift_container_platform/3.9/html/cli_reference/cli-reference-get-started-cli>

   ```bash
   tar xvzf openshift-client-linux.tar.gz
   sudo mv oc /usr/bin/oc

   ```

2. Download and install kubectl

   ```bash
   curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
   chmod +x ./kubectl
   [akiselev@nvfsdn-13 ~]$ sudo mv ./kubectl /usr/local/bin/kubectl
   [akiselev@nvfsdn-13 ~]$ kubectl version --client

   ```

3. Download and install minikube

   ```bash
   curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
   chmod +x minikube
   [akiselev@nvfsdn-13 ~]$ sudo mkdir -p /usr/local/bin/
   [akiselev@nvfsdn-13 ~]$ sudo install minikube /usr/local/bin/
   [akiselev@nvfsdn-13 ~]$ minikube start

   ```

4. Clone and build cnf-certification-test-partner:

   ```git clone git@github.com:redhat-nfvpe/cnf-certification-test-partner.git
   make install```

5. Verify partner and test pods are running: 

   ```bash
   [akiselev@localhost Downloads]$ oc get pods -o wide
   NAME      READY   STATUS    RESTARTS   AGE   IP           NODE       NOMINATED NODE   READINESS GATES
   partner   1/1     Running   0          22m   172.17.0.3   minikube   <none>           <none>
   test      1/1     Running   0          22m   172.17.0.4   minikube   <none>           <none>

   ```

6. Install go:

   ```bash
   wget https://golang.org/dl/go1.15.2.linux-amd64.tar.gz
   tar -C /usr/local -xzf go1.15.2.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin

   ```

7. Clone and build test-network-function:

   ```bash
   git clone git@github.com:redhat-nfvpe/test-network-function.git
   make generic-cnf-tests

   ```

8. If build fails after  
   `go get github.com/onsi/ginkgo/ginkgo`

   Add ginkgo location to the PATH:

   `export PATH=$PATH:~/go/bin`

   Successful generic-cnf-tests run output:

   ```bash
   make generic-cnf-tests -> success

   JUnit report was created: /home/akiselev/test-network-function/test-network-function/cnf-certification-tests_junit.xml

   Ran 4 of 8 Specs in 0.013 seconds
   SUCCESS! -- 4 Passed | 0 Failed | 0 Pending | 4 Skipped
   PASS

   ```
