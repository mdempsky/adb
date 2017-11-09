<template>
  <div id="hot-preview">
    <HotTable :root="root" :settings="hotSettings"></HotTable>
  </div>
</template>

<script>
import HotTable from 'vue-handsontable-official';
import Vue from 'vue';
import {flashMessage} from 'flash_message';

// Constants related to list ordering
// Corresponds to the constants DescOrder and AscOrder in model/activist.go
const DescOrder = 2;
const AscOrder = 1;


function optionsButtonRenderer(instance, td, row, col, prop, value, cellProperties) {
  // SAMER: do something on click?
  td.innerHTML = '<button data-role="trigger" class="btn btn-default"></button>';
  return td;
}

export default {
  name: 'activist-list',
  methods: {
    loadActivists: function() {
      $.ajax({
          url: "/activist/list_range",
          method: "POST",
          data: JSON.stringify(this.pagingParameters),
          success: (data) => {
            var parsed = JSON.parse(data);
            if (parsed.status === "error") {
              flashMessage("Error: " + parsed.message, true);
              return;
            }
            // status === "success"
            var rangedList = parsed.activist_range_list;
            if (rangedList !== null) {
              this.activists = this.activists.concat(rangedList);
              this.pagingParameters.name = rangedList[rangedList.length - 1].name;
            }
          },
          error: () => {
            console.warn(err.responseText);
            flasMessage("Server error: " + err.responseText, true);
          },
        });
    },
    afterChangeCallback: function(change, source) {
      if (source !== 'edit') {
        return;
      }
      var columnIndex = change[0][0];
      var columnName = change[0][1];
      var previousData = change[0][2];
      var newData = change[0][3];

      var activist = this.activists[columnIndex];

      $.ajax({
        url: "/activist/save",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify(activist),
        success: (data) => {
          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          flashMessage(activist.name + " saved");
        },
        error: (err) => {
          console.warn(err.responseText);
          flashMessage("Server error: " + err.responseText, true);
        },
      });
    },

  },
  data: function() {
    return {
      root: 'activists-root',
      activists: [],
      columns: [{
        header: 'Options',
        data: {
          renderer: optionsButtonRenderer,
          readOnly: true,
          disableVisualSelection: true,
        }
      }, {
        header: 'Name',
        data: {
          data: 'name',
        },
      }, {
        header: 'Email',
        data: {
          data: 'email',
        },
      }, {
        header: 'Chapter',
        data: {
          data: 'chapter',
        },
      }, {
        header: 'Phone',
        data: {
          data: 'phone',
        },
      }, {
        header: 'Location',
        data: {
          data: 'location',
        },
      }, {
        header: 'Facebook',
        data: {
          data: 'facebook',
        },
      }, {
        header: 'Core/Staff',
        data: {
          // SAMER: use 'checkbox' instead
          type: 'numeric',
          data: 'core_staff',
        }
      }, {
        header: 'Liberation Pledge',
        data: {
          type: 'numeric',
          data: 'liberation_pledge',
        },
      }, {
        header: 'Global Team Member',
        data: {
          type: 'numeric',
          data: 'global_team_member',
        },
      }, {
        header: 'First Event',
        data: {
          type: 'date',
          data: 'first_event',
          readOnly: true
        },
      }, {
        header: 'Last Event',
        data: {
          type: 'date',
          data: 'last_event',
          readOnly: true,
        },
      }, {
        header: 'Status',
        data: {
          data: 'status',
          readOnly: true,
        }
      }, {
        header: 'Activist Level',
        data: {
          data: 'activist_level',
          readOnly: true,
        }
      }],
      pagingParameters: {
        name: "",
        order: AscOrder,
        limit: 500,
      },
    };
  },
  computed: {
    hotSettings: function() {
      const columns = [];
      const columnHeaders = [];
      for (var i = 0; i < this.columns.length; i++) {
        columns.push(this.columns[i].data);
        columnHeaders.push(this.columns[i].header);
      }
      return {
        data: this.activists,
        columns: columns,
        colHeaders: columnHeaders,
        // SAMER: Disable area selection in general?
        disableVisualSelection: 'area',
        multiSelect: false,
        // SAMER: do we want to enable this?
        fillHandle: false,
        afterChange: this.afterChangeCallback.bind(this),
        // SAMER: not sure if this works
        undo: true,
      };
    },
  },
  created() {
    this.loadActivists();
  },
  components: {
    HotTable,
  },
}

</script>

<style>
</style>
