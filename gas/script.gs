// Function to add a new record to the sheet
function addRecord(sheetName, record) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    sheet = SpreadsheetApp.getActiveSpreadsheet().insertSheet(sheetName);
  }
  sheet.appendRow(record);
}

// Function to read records from the sheet
function readRecords(sheetName) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    Logger.log('Sheet not found');
    return [];
  }
  var data = sheet.getDataRange().getValues();
  return data;
}

// Function to update a record in the sheet
function updateRecord(sheetName, rowNum, colNum, newValue) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    Logger.log('Sheet not found');
    return;
  }
  sheet.getRange(rowNum, colNum).setValue(newValue);
}

// Function to delete a record from the sheet
function deleteRecord(sheetName, rowNum) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    Logger.log('Sheet not found');
    return;
  }
  sheet.deleteRow(rowNum);
}

// Function to search records in the sheet
function searchRecords(sheetName, searchColumn, searchValue) {
  var sheet = SpreadsheetApp.getActiveSpreadsheet().getSheetByName(sheetName);
  if (!sheet) {
    Logger.log('Sheet not found');
    return [];
  }
  var data = sheet.getDataRange().getValues();
  var results = [];
  
  // Determine the column index based on the search column name
  var columnIndex = -1;
  if (searchColumn === 'email') {
    columnIndex = 0;
  } else if (searchColumn === 'name') {
    columnIndex = 1;
  } else if (searchColumn === 'studentId') {
    columnIndex = 2;
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
  
  if (action === 'read') {
    response = readRecords(sheetName);
  } else if (action === 'search') {
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
  } else if (action === 'update') {
    updateRecord(sheetName, params.rowNum, params.colNum, params.newValue);
    response = { success: true };
  } else if (action === 'delete') {
    deleteRecord(sheetName, params.rowNum);
    response = { success: true };
  } else if (action === 'search') {
    response = searchRecords(sheetName, params.searchColumn, params.searchValue);
  } else {
    response = { error: 'Invalid action' };
  }
  
  return ContentService.createTextOutput(JSON.stringify(response)).setMimeType(ContentService.MimeType.JSON);
}
