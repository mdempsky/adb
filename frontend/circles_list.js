import CirclesList from 'CirclesList.vue';
import Vue from 'vue';

export function initializeApp() {
  var vm = new Vue({
    el: "#app",
    render: function(h) {
      return h(CirclesList);
    }
  });
}
