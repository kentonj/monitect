<template lang="pug">
.container
  .box
    .title {{ sensor.name || '' }}
    .box
      p {{ lastSensorReading.value }}
</template>

<script>
import axios from 'axios';

function getSensorReadings(sensorId, limit) {
  return axios
    .get(`/api/sensors/${sensorId}/readings`, { limit })
    .then((response) => response.data.sensorReadings);
}

export default {
  name: 'Sensor',
  props: {
    sensor: Object,
  },
  data() {
    return {
      sensorReadings: [],
      lastSensorReading: {},
    };
  },
  methods: {
    setSensorReadings() {
      getSensorReadings(this.$props.sensor.id, 10).then((sensorReadings) => {
        this.sensorReadings = sensorReadings;
        this.lastSensorReading = sensorReadings[sensorReadings.length - 1];
      });
    },
    pollForSensorReadings() {
      // fetch the initial sensor readings
      this.setSensorReadings();
      // look for a new set of sensor readings regularly
      setInterval(() => {
        this.setSensorReadings();
      }, 10000);
    },
  },
  mounted() {
    this.pollForSensorReadings();
  },
};
</script>
