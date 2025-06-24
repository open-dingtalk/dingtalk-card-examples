import json


def convert_json_values_to_string(obj: dict) -> str:
    """
    Dump the attributes of a dictionary to a string.
    """
    result = {}
    for key, value in obj.items():
        if isinstance(value, str):
            result[key] = value
        else:
            result[key] = json.dumps(value)
    return result
