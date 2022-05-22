<template lang="pug">
.container
  .title.is-1 {{ title }}
  .columns
    .column
      Camera(v-for="camera in cameras"
      :key="camera.id"
      :camera="camera"
      )
    .column
      Sensor(v-for="sensor in sensors"
      :key="sensor.id"
      :sensor="sensor"
      )
</template>

<script>
import Sensor from '@/components/Sensor.vue';
import Camera from '@/components/Camera.vue';
import axios from 'axios';

function getSensors() {
  return axios.get('/api/sensors').then((response) => response.data.sensors);
}

export default {
  name: 'Dashboard',
  components: {
    Sensor,
    Camera,
  },
  data() {
    return {
      title: 'monitect',
      sensors: [],
      cameras: [],
    };
  },
  methods: {
    // get and set both cameras and sensors
    setSensors() {
      getSensors().then((sensors) => {
        const cameras = sensors.filter((sensor) => sensor.type === 'camera');
        const nonCameras = sensors.filter((sensor) => sensor.type !== 'camera');
        this.cameras = cameras;
        this.sensors = nonCameras;
      });
    },
  },
  created() {
    this.setSensors();
  },
};
</script>
