# dns3l instance tests

The instance tests can be used to test the most common API calls of a DNS3L instance.
Note that the test modifies the instance (it claims a cert, deletes any possible
conflicting certs, and deletes the created cert afterward), so keep caution when running it.

Usage:

```bash
export DNS3L_INSTANCE=https://dns3l.example.com
export DNS3L_TESTCA=les
export DNS3L_TESTCERT=footest.example.com
export DNS3L_USER=<username>
export DNS3L_PASS=<password>

go run test/instance/main.go

```