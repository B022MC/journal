import importlib.util
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT_PATH = Path(__file__).resolve().with_name("render_remote_journal_configs.py")
sys.path.insert(0, str(SCRIPT_PATH.parent))
SPEC = importlib.util.spec_from_file_location("render_remote_journal_configs", SCRIPT_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = MODULE
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


class RenderRemoteJournalConfigsTest(unittest.TestCase):
    def test_rewrite_db_block_replaces_split_settings(self) -> None:
        source = """DB:
  ReadWriteSplit: true
  DataSource: "journal:local@tcp(127.0.0.1:13306)/journal"
  Policy: roundRobin
  Replicas:
    - "journal:local@tcp(127.0.0.1:13307)/journal"
    - "journal:local@tcp(127.0.0.1:13308)/journal"
  Cache:
    - user
Other:
  Keep: true
"""
        rendered = MODULE.rewrite_db_block(
            source,
            "journal:redacted@tcp(remote-host:3306)/journal?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
        )

        self.assertIn('DataSource: "journal:redacted@tcp(remote-host:3306)/journal?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai"', rendered)
        self.assertIn("ReadWriteSplit: false", rendered)
        self.assertIn("  Cache:", rendered)
        self.assertIn("Other:", rendered)
        self.assertNotIn("ReadWriteSplit: true", rendered)
        self.assertNotIn("Policy:", rendered)
        self.assertNotIn("Replicas:", rendered)
        self.assertNotIn("127.0.0.1:13307", rendered)
        self.assertNotIn("127.0.0.1:13308", rendered)

    def test_rewrite_redis_pass_updates_only_redis_block(self) -> None:
        source = """Redis:
  Host: 127.0.0.1:16379
  Pass: "old-pass"
CacheRedis:
  Host: 127.0.0.1:16379
  Pass: "old-cache-pass"
Telemetry:
  Endpoint: http://127.0.0.1:4318
"""
        rendered = MODULE.rewrite_redis_pass(source, "new-pass")

        self.assertEqual(rendered.count('Pass: "new-pass"'), 2)
        self.assertNotIn('Pass: "old-pass"', rendered)
        self.assertNotIn('Pass: "old-cache-pass"', rendered)
        self.assertIn("Telemetry:", rendered)

    def test_rewrite_etcd_keys_suffixes_nested_and_top_level_keys(self) -> None:
        source = """Etcd:
  Hosts:
    - 127.0.0.1:12379
  Key: user.rpc
PaperRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:12379
    Key: paper.rpc
"""
        rendered = MODULE.rewrite_etcd_keys(source, ".b022mc.20260314")

        self.assertIn("Key: user.rpc.b022mc.20260314", rendered)
        self.assertIn("Key: paper.rpc.b022mc.20260314", rendered)
        self.assertEqual(rendered.count(".b022mc.20260314"), 2)

    def test_default_etcd_key_suffix_uses_owner_and_date(self) -> None:
        suffix = MODULE.default_etcd_key_suffix("B022 MC", "2026-03-14")

        self.assertEqual(suffix, ".b022-mc.20260314")

    def test_rewrite_rpc_server_to_direct_removes_top_level_etcd(self) -> None:
        source = """Name: user.rpc
ListenOn: 0.0.0.0:9001
Etcd:
  Hosts:
    - 127.0.0.1:12379
  Key: user.rpc
DB:
  ReadWriteSplit: false
"""
        rendered = MODULE.rewrite_rpc_server_to_direct(source)

        self.assertNotIn("Etcd:", rendered)
        self.assertNotIn("Key:", rendered)
        self.assertIn("ListenOn: 0.0.0.0:9001", rendered)

    def test_rewrite_rpc_clients_to_direct_replaces_etcd_with_endpoints(self) -> None:
        source = """UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:12379
    Key: user.rpc
  Timeout: 3000
  NonBlock: true
"""
        rendered = MODULE.rewrite_rpc_clients_to_direct(source)

        self.assertIn("  Endpoints:", rendered)
        self.assertIn("    - 127.0.0.1:9001", rendered)
        self.assertIn("  Timeout: 3000", rendered)
        self.assertNotIn("Etcd:", rendered)
        self.assertNotIn("Key:", rendered)

    def test_purge_generated_outputs_removes_known_targets_only(self) -> None:
        with tempfile.TemporaryDirectory() as temp_dir:
            output_root = Path(temp_dir)
            known_target = output_root / MODULE.SERVICE_CONFIGS[0].output
            known_target.parent.mkdir(parents=True, exist_ok=True)
            known_target.write_text("stale", encoding="utf-8")

            unrelated = output_root / "notes.txt"
            unrelated.write_text("keep", encoding="utf-8")

            MODULE.purge_generated_outputs(output_root)

            self.assertFalse(known_target.exists())
            self.assertEqual(unrelated.read_text(encoding="utf-8"), "keep")

    def test_render_config_writes_remote_yaml(self) -> None:
        with tempfile.TemporaryDirectory() as backend_dir, tempfile.TemporaryDirectory() as output_dir:
            backend_root = Path(backend_dir)
            source = backend_root / MODULE.SERVICE_CONFIGS[0].source
            source.parent.mkdir(parents=True, exist_ok=True)
            source.write_text(
                """Name: journal-api
DB:
  ReadWriteSplit: true
  DataSource: "journal:local@tcp(127.0.0.1:13306)/journal"
  Policy: roundRobin
  Replicas:
    - "journal:local@tcp(127.0.0.1:13307)/journal"
Redis:
  Host: 127.0.0.1:16379
  Pass: "local-pass"
UserRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:12379
    Key: user.rpc
Other:
  Keep: true
""",
                encoding="utf-8",
            )

            rendered_path = MODULE.render_config(
                backend_root,
                MODULE.SERVICE_CONFIGS[0],
                Path(output_dir),
                "journal:redacted@tcp(remote-host:3306)/journal?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai",
                "remote-redis-pass",
                ".b022mc.20260314",
                "direct",
            )
            rendered = rendered_path.read_text(encoding="utf-8")

            self.assertTrue(rendered_path.exists())
            self.assertIn("Name: journal-api", rendered)
            self.assertIn("ReadWriteSplit: false", rendered)
            self.assertIn("Other:", rendered)
            self.assertIn('Pass: "remote-redis-pass"', rendered)
            self.assertIn("Endpoints:", rendered)
            self.assertIn("127.0.0.1:9001", rendered)
            self.assertNotIn("Etcd:", rendered)
            self.assertNotIn("127.0.0.1:13306", rendered)
            self.assertNotIn("Replicas:", rendered)


if __name__ == "__main__":
    unittest.main()
