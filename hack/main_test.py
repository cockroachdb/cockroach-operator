import platform
import pytest
import sys

class TestPythonVersion:
    def test_version(self):
        assert('bazel-out/k8-fastbuild/bin/main_test.runfiles/python_interpreter/python_bin' in sys.executable)
        assert(platform.python_version() == "3.8.3")

if __name__ == "__main__":
    import pytest
    raise SystemExit(pytest.main([__file__]))
