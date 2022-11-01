import axios from 'axios';

const wsProto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsBaseUrl = `${wsProto}//${window.location.host}/api`;
const httpBaseUrl = `${window.location.protocol}//${window.location.host}/api`;

function getSensors() {
  const url = `${httpBaseUrl}/sensors`;
  console.log(`requesting ${url}`);
  return axios.get(url).then((response) => response.data.sensors);
}

function readSensorSocket(sensorId, clientId) {
  const wsUrl = `${wsBaseUrl}/sensors/${sensorId}/read?clientId=${clientId}`;
  console.log(`connecting to socket ${wsUrl}`);
  return new WebSocket(wsUrl);
}

export default {
  readSensorSocket,
  getSensors,
};
