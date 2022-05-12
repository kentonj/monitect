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
      getSensorReadings(this.$props.sensor.id).then((sensorReadings) => {
        this.sensorReadings = sensorReadings;
        this.lastSensorReading = sensorReadings[sensorReadings.length - 1];
      });
    },
  },
  created() {
    this.setSensorReadings();
  },
};
</script>
