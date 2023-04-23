const page_to_name = {
  'home': 'Instructions',
  'tutorial1': 'Tutorial 1',
  'tutorial2': 'Tutorial 2',
  'tutorial3': 'Tutorial 3',
  'level1':    'Level 1',
  'level2':    'Level 2',
  'level3':    'Level 3',
};

const target_map = {
  'print_flag()': 'level1',
  'return_flag()': 'level2',
  'write_flag_to_file()': 'level3',
}

function showError(msg) {
  console.error(msg);
  $('#error').text(msg);
  $('#error').removeClass('d-none');
  $('#error').get(0).scrollIntoView({ behavior: 'smooth' });

  // Hide it after a bit
  setTimeout(() => $('#error').addClass('d-none'), 5000);
}

function addGadget(section, element) {
  let li = $('<li class="nav-item">');
  element.addClass('nav-link');
  li.append(element);
  $(`#${section}`).append(li);
}

function handleExecuteResponse(response) {
  console.log(response);
  if(response['error']) {
    showError(`Something went wrong executing: ${ response['error'] }`);
    return;
  }

  $('#exit-reason').text(response.exit_reason);

  if(!response.stdout || response.stdout === '') {
    $('#stdout').text('n/a');
  } else {
    $('#stdout').text(response.stdout);
  }

  $('#instructions').empty();
  response.history.forEach((instruction) => {
    let tr = $(`<tr>`);
    tr.append($('<td>').text(instruction.instruction));
    $('#instructions').append(tr);
  });

  $('#output').show();
  $('#output').get(0).scrollIntoView({ behavior: 'smooth' });
}

function loadGadgets() {
  $('#target-function').empty();
  $('#gadgets').empty();
  $('#functions').empty();
  $('#stack').empty();

  let stack = [];

  let getStack = () => {
    return stack.map((s) => s.hex).join('')
  };

  const saveState = () => {
    let json = JSON.stringify(stack);
    window.localStorage.setItem('code', json);
    console.log(`Saved: ${ json }`);
  };

  let regenerateStack = () => {
    $('#stack').empty();
    let lastTr;

    stack.forEach((s, i) => {
      let tr = $(`<tr title="${s.description} (click to remove!)">`);
      tr.append($('<td>').text(s.hex));
      tr.append($('<td>').text(s.address));
      tr.append($('<td>').text(s.name));
      $('#stack').append(tr);
      lastTr = tr;

      tr.click(() => {
        // This bit of logic make sure you can't remove a pop or constant separately (easily)
        if(s.connected_to_last_value) {
          stack.splice(i - 1, 2);
        } else if(s.connected_to_next_value) {
          stack.splice(i, 2);
        } else {
          stack.splice(i, 1);
        }
        regenerateStack();
      });
    });

    if(lastTr) {
      lastTr.get(0).scrollIntoView({ behavior: 'smooth' });
    }

    // This seems like a fine place to save it
    saveState();
  };

  // Load the stack from localStorage
  const restoreState = () => {
    let json = window.localStorage.getItem('code');
    if(!json) {
      console.log("No state to restore!");
      return;
    }

    stack = JSON.parse(json);
    console.log(`Loaded stack from localStorage: ${ json }`);
    regenerateStack();
  };

  $.getJSON('/gadgets', (data) => {
    data.forEach((gadget) => {
      let e = $(`<a class="gadget" title="${ gadget.description }">${ gadget.name }</a>`)

      // A bit of logic to hide/show the correct target function
      let associatedLevel = target_map[gadget.name];
      if(associatedLevel) {
        e.addClass('sometimes-hidden');
        e.addClass(associatedLevel);
        e.hide();
      }

      e.click(() => {
        if(gadget.prompt_for_integer) {
          let value = prompt(`Please enter an integer to \"${ gadget.name }\" into! (Use 0x prefix for hex)`);
          if(value === null || value === '') {
            return;
          }

          let intValue = parseInt(value);
          let tempValue = intValue;
          let hexValue = '';

          for(let i = 0; i < 8; i++) {
            hexValue += ('0' + (tempValue & 0xFF).toString(16)).slice(-2);
            tempValue >>= 8;
          }

          // Push it on the stack like a gadget
          stack.push({
            connected_to_next_value: true,
            ...gadget
          });

          stack.push({
            name: `Constant (consumed by pop): ${value}`,
            address: `0x${ intValue.toString(16) }`,
            description: `A constant value (${ intValue })`,
            hex: hexValue,
            connected_to_last_value: true,
          });
        } else {
          stack.push(gadget);
        }
        regenerateStack();
      });

      addGadget(gadget.type, e);
    });

    // Show/hide depending on the page
    doChangePage();
    restoreState();
  });

  $('#execute').click(() => {
    $('#output').show();
    $('#output-loading').show();
    $('#output-output').hide();
    $.post(
      "/execute",
      JSON.stringify({ code: getStack() }),
      (response) => {
        $('#output-loading').hide();
        $('#output-output').show();
        handleExecuteResponse(response);
      },
    );
  });

  $('#clear').click(() => {
    stack = [];
    regenerateStack();
  });
}

function doChangePage() {
  $('#temp-page-title').hide();

  // Hide all the stuff that gets hidden
  $('.sometimes-hidden').hide();

  let page = window.location.hash.substr(1);

  if(!page || page === '') {
    page = 'home';
  }
  let name = page_to_name[page];

  if(!name) {
    showError("Invalid page id!");
    return;
  }


  // Do the three target functions specially

  // Show the stuff associated with the current page
  $(`.${ page }`).show();

  // If we're not on home, show the not-home stuff
  if(page !== 'home') {
    $('.not-home').show();
  }
}

$(window).on("hashchange", () => {
  doChangePage();
});

$(document).ready(() => {
  // This will call doChangePage after loading
  loadGadgets();
});
