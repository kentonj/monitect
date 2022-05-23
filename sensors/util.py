import requests


def get_sensor(host: str, type_match: str, name_match: str):
    """Check for a 'camera' type sensor with this name, if not, create it, otherwise use the existing camera."""
    resp = requests.get(f'{host}/api/sensors')
    resp.raise_for_status()
    for sensor in resp.json()['sensors']:
        if sensor['type'] == type_match and sensor['name'] == name_match:
            return sensor
    else:
        return None
