<template lang="pug">
.container
  .box
    .title.is-3 {{ sensor.name || '' }}
    .box
      .title.is-4 {{ latestSensorReading.value }} ({{ sensor.unit || '' }})
      .subtitle.is-6 {{ latestSensorReading.createdTime }}
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
      latestSensorReading: {},
    };
  },
  methods: {
    setSensorReadings() {
      getSensorReadings(this.$props.sensor.id, 10).then((readings) => {
        this.sensorReadings = readings;
        // get the latest sensorReading
        const latestSensorReading = readings.reduce((p, c) => {
          const comp = p.createdAt > c.createdAt;
          return comp ? p : c;
        });
        const createdAt = new Date(Date.parse(latestSensorReading.createdAt));
        latestSensorReading.createdTime = createdAt.toLocaleTimeString();
        this.latestSensorReading = latestSensorReading;
        console.log(`new latest sensor reading: ${this.latestSensorReading.value}`);
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
