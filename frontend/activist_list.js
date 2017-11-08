import ActivistList2 from 'ActivistList2.vue';
import Vue from 'vue';

export function initializeApp() {
  var vm = new Vue({
    el: "#app",
    render: function(h) {
      return h(ActivistList2);
    }
  });
}
