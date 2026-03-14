import importlib.util
import sys
import unittest
from pathlib import Path


SCRIPT_PATH = Path(__file__).resolve().with_name("check_remote_journal_rollback_dry_run.py")
SPEC = importlib.util.spec_from_file_location("check_remote_journal_rollback_dry_run", SCRIPT_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = MODULE
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


REMOTE_OUTPUT = """[remote] rendering configs into /tmp/journal-remote-validation
[infra] starting docker-compose services: redis etcd jaeger
[dry-run] user-rpc: go run rpc/user/user.go -f /tmp/journal-remote-validation/rpc/user/etc/user.remote.yaml
[dry-run] admin-api: go run admin-api/admin.go -f /tmp/journal-remote-validation/admin-api/etc/admin-api.remote.yaml
Profile:
  - mode:          dev
  - config profile:remote
Infra:
  - redis:          16379
  - etcd:           12379
  - jaeger ui:      16686
"""


LOCAL_OUTPUT = """[infra] starting docker-compose services: mysql-master mysql-replica1 mysql-replica2 redis etcd jaeger prometheus grafana
[dry-run] user-rpc: go run rpc/user/user.go -f rpc/user/etc/user.yaml
[dry-run] admin-api: go run admin-api/admin.go -f admin-api/etc/admin-api.yaml
[dry-run] cron: go run cmd/cron/main.go
Profile:
  - mode:          dev
  - config profile:local
Infra:
  - mysql-master:   13306
  - mysql-replica1: 13307
  - mysql-replica2: 13308
  - grafana:        13000
"""


class CheckRemoteJournalRollbackDryRunTest(unittest.TestCase):
    def test_validate_remote_output_accepts_expected_markers(self) -> None:
        self.assertEqual(MODULE.validate_remote_output(REMOTE_OUTPUT), [])

    def test_validate_remote_output_rejects_local_artifacts(self) -> None:
        output = REMOTE_OUTPUT + "\n[dry-run] cron: go run cmd/cron/main.go\n"
        errors = MODULE.validate_remote_output(output)

        self.assertIn("forbidden:cmd/cron/main.go", errors)

    def test_validate_local_output_accepts_expected_markers(self) -> None:
        self.assertEqual(MODULE.validate_local_output(LOCAL_OUTPUT), [])

    def test_validate_local_output_rejects_remote_overlay(self) -> None:
        output = LOCAL_OUTPUT + "\n[dry-run] user-rpc: go run rpc/user/user.go -f /tmp/journal-remote-validation/rpc/user/etc/user.remote.yaml\n"
        errors = MODULE.validate_local_output(output)

        self.assertIn("forbidden:/tmp/journal-remote-validation", errors)


if __name__ == "__main__":
    unittest.main()
