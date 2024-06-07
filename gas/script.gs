// Function to get sheet
function getSheet(sheetName) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    sheet = SpreadsheetApp.getActiveSpreadsheet().insertSheet(sheetName);
  }
  return sheet;
}

// Function to get column index by given column
function getColumnIndex(sheet, column) {
  var data = sheet.getDataRange().getValues();
  for (var i = 0; i < data.length; i++) {
    if (data[0][i] == column) return i;
  }
}

// Function to get row on sheet (which is added by 1)
function getRow(sheet, column, value) {
  var col = getColumnIndex(sheet, column);
  var data = sheet.getDataRange().getValues();
  for (var i = 1; i < data.length; i++) {
    if (data[i][col] == value) return i + 1;
  }
  return -1;
}

// Function to add a new record to the sheet
function addRecord(sheetName, record) {
  var sheet = getSheet(sheetName);
  sheet.appendRow(record);
}

// Function to delete a record from the sheet
function deleteRecord(sheetName, columnName, rowValue) {
  var sheet = getSheet(sheetName);
  var rowNum = getRow(sheet, columnName, rowValue);
  sheet.deleteRow(rowNum);
}

// Function to search records in the sheet
function searchRecords(sheetName, searchColumn, searchValue) {
  var sheet = getSheet(sheetName);
  var data = sheet.getDataRange().getValues();
  var results = [];
  
  // Determine the column index based on the search column name
  var columnIndex = getColumnIndex(sheet, searchColumn);
  if (columnIndex == -1) {
    Logger.log('Sheetname: %s with searchColumn: %s not found', sheetName, searchColumn)
    return []
  }
  
  // Iterate through data to find matching records
  for (var i = 1; i < data.length; i++) { // Skip header row (i = 1)
    if (data[i][columnIndex] && data[i][columnIndex].toString().toLowerCase() === searchValue.toLowerCase()) {
      results.push(data[i]);
    }
  }
  return results;
}

// Handle GET requests
function doGet(e) {
  var sheetName = e.parameter.sheetName;
  var action = e.parameter.action;
  var response;
  
  if (action === 'search') {
    var searchColumn = e.parameter.searchColumn;
    var searchValue = e.parameter.searchValue;
    response = searchRecords(sheetName, searchColumn, searchValue);
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
