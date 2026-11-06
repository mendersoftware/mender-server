import pytest
from mender_client.models.deployment_v1 import DeploymentV1


# The validators can only be generated when they are used,
# and I decided to use and test them via the DeploymentV1 schema.
class TestCommonValidation:
    @pytest.mark.parametrize(
        "value",
        [
            "",
            "Abc",
            "._-",
        ],
    )
    def test_identifier_ok(self, value):
        assert DeploymentV1.id_validate_regular_expression(value) == value

    @pytest.mark.parametrize(
        "value",
        [
            "id space",
            "id\nnewline",
            "id\ttab",
            "id@",
            "id/",
            "id$",
        ],
    )
    def test_identifier_invalid(self, value):
        with pytest.raises(ValueError, match="must validate the regular expression"):
            DeploymentV1.id_validate_regular_expression(value)

    @pytest.mark.parametrize(
        "value",
        [
            "",
            "A",
            "Name 123!@#-_.",
            "عليك تجربة الفالدي 😉",
        ],
    )
    def test_name_ok(self, value):
        assert DeploymentV1.name_validate_regular_expression(value) == value

    @pytest.mark.parametrize(
        "value",
        [
            "name\nnewline",
            "name\ttab",
            "name\x00null",
        ],
    )
    def test_name_invalid(self, value):
        with pytest.raises(ValueError, match="must validate the regular expression"):
            DeploymentV1.name_validate_regular_expression(value)

    @pytest.mark.parametrize(
        "value",
        [
            "",
            "A",
            "text 123!@#-_.",
            "text\nnewline",
            "text\ttab",
            "text\rcarriage",
            "هل تعلم ان لدينا ميزة سجلات المراجعة؟ جرب الانتربرايز 👀",
        ],
    )
    def test_text_ok(self, value):
        assert DeploymentV1.artifact_name_validate_regular_expression(value) == value

    @pytest.mark.parametrize(
        "value",
        [
            "text\x00null",
            "text\x0bverticalTab",
            "text\x0eshiftOut",
        ],
    )
    def test_text_invalid(self, value):
        with pytest.raises(ValueError, match="must validate the regular expression"):
            DeploymentV1.artifact_name_validate_regular_expression(value)
