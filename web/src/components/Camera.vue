<template lang="pug">
.container
  .box
    .title {{ camera.name || '' }}
    .box
      img(:src="`data:image/png;base64,${imageBase64}`")
</template>

<script>
import axios from 'axios';

function getLatestImage(sensorId) {
  return axios
    .get(`/api/sensors/${sensorId}/images/latest`, { responseType: 'arraybuffer' })
    .then((response) => {
      const buffer = Buffer.from(response.data, 'binary');
      return buffer.toString('base64');
    });
}

export default {
  name: 'Camera',
  props: {
    camera: Object,
  },
  data() {
    return {
      imageBase64: '',
    };
  },
  methods: {
    setLatestImage() {
      getLatestImage(this.$props.camera.id).then((imageBase64) => {
        this.imageBase64 = imageBase64;
      });
    },
    pollForImages() {
      // fetch the initial image
      this.setLatestImage();
      // look for a new one regularly
      setInterval(() => {
        this.setLatestImage();
      }, 3000);
    },
  },
  mounted() {
    this.pollForImages();
  },
};
</script>
