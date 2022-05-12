<template>
  <div class="home">
    <img alt="Monitect Logo" src="../assets/logo.png">
    <Dashboard/>
  </div>
</template>

<script>
// @ is an alias to /src
import Dashboard from '@/components/Dashboard.vue';

export default {
  name: 'Home',
  components: {
    Dashboard,
  },
};
</script>
