<template>
  <div id="app" class="main">
    <button @click="blah">yo</button>
    <HotTable :root="root" :settings="hotSettings" :data="activists"></HotTable>
    <modal
       name="activist-options-modal"
       height="auto"
       classes="no-background-color no-top"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">{{currentActivist.name}}</h2>
          </div>
          <div class="modal-body">
            <ul class="activist-options-body">
              <li>
                <a @click="showModal('merge-activist-modal', currentActivist, activistIndex)">Merge Activist</a>
              </li>
              <li>
                <a @click="showModal('hide-activist-modal', currentActivist, activistIndex)">Hide Activist</a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </modal>
    <modal
       name="merge-activist-modal"
       :height="650"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Merge activist</h2>
          </div>
          <div class="modal-body">
            <p>Merging activists is used to combine redundant activist entries</p>
            <p>
              Merging this activist does two things:
            </p>
            <ul>
              <li>all of this activist&#39;s attendance data will be merged into the target activist</li>
              <li>this activist will be hidden</li>
            </ul>
            <p>
              Non-attendance data (e.g. email, location, etc) is <strong>NOT</strong> merged.
            </p>
            <p>Merge {{currentActivist.name}} into another activist:</p>
            <p>
              Target activist: <select id="merge-target-activist" class="filter-margin" style="min-width: 200px"></select>
            </p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-danger" @click="confirmMergeActivistModal" v-focus>Merge activist</button>
          </div>
        </div>
      </div>
    </modal>
    <modal
       name="hide-activist-modal"
       :height="400"
       classes="no-background-color"
       @opened="modalOpened"
       @closed="modalClosed"
       >
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h2 class="modal-title">Hide activist</h2>
          </div>
          <div class="modal-body">
            <p>Are you sure you want to hide {{currentActivist.name}}?</p>
            <p>Hiding an activist hides them from the activist list page but does not delete any event data associated with them. If this activist is a duplicate of another activist, you should merge them instead.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" @click="hideModal">Close</button>
            <button type="button" v-bind:disabled="disableConfirmButton" class="btn btn-danger" @click="confirmHideActivistModal" v-focus>Hide activist</button>
          </div>
        </div>
      </div>
    </modal>
  </div>
</template>

<script>
import vmodal from 'vue-js-modal';
import HotTable from 'external/vue-handsontable-official/HotTable.vue';
import Vue from 'vue';
import {focus} from 'directives/focus';
import {flashMessage} from 'flash_message';
import {EventBus} from 'EventBus';
import {initActivistSelect} from 'chosen_utils';

Vue.use(vmodal);

// Constants related to list ordering
// Corresponds to the constants DescOrder and AscOrder in model/activist.go
const DescOrder = 2;
const AscOrder = 1;

window.showOptionsModal = function (row) {
  EventBus.$emit('activist-show-options-modal', row);
}

function optionsButtonRenderer(instance, td, row, col, prop, value, cellProperties) {
  td.innerHTML = '<button ' +
    'data-role="trigger" ' +
    'class="btn btn-default btn-xs dropdown-toggle glyphicon glyphicon-option-horizontal" ' +
    'type="button" ' +
    'onclick="window.showOptionsModal(' + row + ')"></button>';
  return td;
}

const component = {
  name: 'activist-list',
  methods: {
    blah: function() {
      console.log(this.activists);
    },
    showOptionsModal: function(row) {
      var activist = this.activists[row];
      this.showModal('activist-options-modal', activist, row);
    },
    showModal: function(modalName, activist, index) {
      // Check to see if there's a modal open, and close it if so.
      if (this.currentModalName) {
        this.hideModal();
      }

      // SAMER: why nexttick
      Vue.nextTick(() => {
        this.currentActivist = activist;

        if (index != undefined) {
          this.activistIndex = index; // needed for updating activist
        } else {
          this.activistIndex = -1;
        }

        this.currentModalName = modalName;
        this.$modal.show(modalName);
      });
    },
    hideModal: function() {
      if (this.currentModalName) {
        this.$modal.hide(this.currentModalName);
      }
      this.currentModalName = '';
      this.activistIndex = -1;
      this.currentActivist = {};
    },
    modalOpened: function() {
      // Add noscroll to body tag so it doesn't scroll while the modal
      // is shown.
      $(document.body).addClass('noscroll');
      this.disableConfirmButton = false;

      if (this.currentModalName == "merge-activist-modal") {
        // For some reason, even though this function is supposed to
        // fire after the modal is visible on the dom, the modal isn't
        // there. Vue.nextTick doesn't work for some reason, so we're
        // just going to keep calling setTimeout until the modal shows
        // up.
        var interval;
        var fn = () => {
          if ($('#merge-target-activist')[0]) {
            clearInterval(interval);
            initActivistSelect('#merge-target-activist', this.currentActivist.name);
          }
        };
        interval = setInterval(fn, 50);
      }

    },
    modalClosed: function() {
      // Allow body to scroll after modal is closed.
      $(document.body).removeClass('noscroll');
    },
    confirmMergeActivistModal: function() {
      var targetActivistName = $("#merge-target-activist").val();
      if (!targetActivistName) {
        flashMessage("Must choose an activist to merge into", true);
        return;
      }

      this.disableConfirmButton = true;

      $.ajax({
        url: "/activist/merge",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({
          current_activist_id: this.currentActivist.id,
          target_activist_name: targetActivistName,
        }),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          flashMessage(this.currentActivist.name + " was merged into " + targetActivistName);

          // Remove activist from list.
          this.activists = this.activists.slice(0, this.activistIndex).concat(
            this.activists.slice(this.activistIndex+1));
          this.hideModal();
        },
        error: () => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage("Server error: " + err.responseText, true);
        },
      });
    },
    confirmHideActivistModal: function() {
      this.disableConfirmButton = true;

      $.ajax({
        url: "/activist/hide",
        method: "POST",
        contentType: "application/json",
        data: JSON.stringify({id: this.currentActivist.id}),
        success: (data) => {
          this.disableConfirmButton = false;

          var parsed = JSON.parse(data);
          if (parsed.status === "error") {
            flashMessage("Error: " + parsed.message, true);
            return;
          }
          flashMessage(this.currentActivist.name + " was hidden");

          // Remove activist from list.
          this.activists = this.activists.slice(0, this.activistIndex).concat(
            this.activists.slice(this.activistIndex+1));

          console.log(this.activists);
          this.hideModal();
        },
        error: () => {
          this.disableConfirmButton = false;

          console.warn(err.responseText);
          flashMessage("Server error: " + err.responseText, true);
        },
      });
    },
    loadActivists: function() {
      console.log("loadActivists", this);
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
      if (source !== 'edit' &&
          source !== 'UndoRedo.undo' &&
          source !== 'UndoRedo.redo') {
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
      currentModalName: '',
      activistIndex: -1,
      currentActivist: {},
      disableConfirmButton: false,
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
    EventBus.$on('activist-show-options-modal', (row) => {
      this.showOptionsModal(row);
    });
  },
  components: {
    HotTable,
  },
  directives: {
    focus,
  },
};

export default component;

</script>

<style>
  .activist-options-body a {
    color: #337ab7;
    cursor: pointer;
  }
</style>
