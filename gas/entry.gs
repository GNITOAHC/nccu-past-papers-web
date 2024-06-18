// Handle GET requests
function doGet(e) {
  var sheetName = e.parameter.sheetName;
  var action = e.parameter.action;
  var response;

  if (action === 'search') {
    var searchColumn = e.parameter.searchColumn;
    var searchValue = e.parameter.searchValue;
    response = searchRecords(sheetName, searchColumn, searchValue);
  } else if (action === 'readall') {
    response = getAll(sheetName);
  } else {
    response = { error: 'Invalid action' };
  }

  return ContentService.createTextOutput(JSON.stringify(response)).setMimeType(ContentService.MimeType.JSON);
}

// Handle POST requests
function doPost(e) {
  var params = JSON.parse(e.postData.contents);
  var sheetName = params.sheetName;
  var action = params.action;
  var response;

  if (action === 'add') {
    addRecord(sheetName, params.record);
    response = { success: true };
  } else if (action === 'delete') {
    deleteRecord(sheetName, params.columnName, params.rowValue);
    response = { success: true };
  } else {
    response = { error: 'Invalid action' };
  }

  return ContentService.createTextOutput(JSON.stringify(response)).setMimeType(ContentService.MimeType.JSON);
}