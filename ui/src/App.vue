<template>
  <div id="app">

    <div class="ui search">
      <div class="ui icon input">
        <input class="prompt" type="text" placeholder="Queue name">
        <i class="search icon"></i>
      </div>
    </div>

    <!-- <img src="./assets/logo.png"> -->
    <table class="ui fixed single line blue table">
      <thead>
        <tr>
          <th class="center aligned">Publisher</th>
          <th>Queue</th>
          <th class="center aligned">Consumer</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in items" v-bind:key="item.queue">
          <td class="center aligned">
            <div class="ui toggle checkbox">
              <input type="checkbox" name="publisher">
              <label></label>
            </div>
          </td>
          <td>{{item.queue}}</td>
          <td class="center aligned">
            <div class="ui toggle checkbox">
              <input type="checkbox" name="consumer">
              <label></label>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script>
export default {
  name: 'app',
  data() {
    return {
      msg: 'RabbitRunr',
      items: [
        { queue: '743afbc7-9208-40a8-8d22-1c33c3d6e6b0' },
        { queue: '41766c93-aafe-44ce-9e17-26b0f68caa0e' }
      ]
    }
  }
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}

h1,
h2 {
  font-weight: normal;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  display: inline-block;
  margin: 0 10px;
}

a {
  color: #42b983;
}
</style>
