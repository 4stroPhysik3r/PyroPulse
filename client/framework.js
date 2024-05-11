// DOM MANIPULATION
// Function to create a DOM element with optional attributes and children
function createElement(tag, attributes = {}, children = []) {
  const element = document.createElement(tag);

  // Set attributes
  for (const [key, value] of Object.entries(attributes)) {
    element.setAttribute(key, value);
  }

  // Append children
  children.forEach((child) => {
    if (typeof child === "string") {
      appendChild(element, document.createTextNode(child));
    } else {
      appendChild(element, child);
    }
  });

  return element;
}

// Function to append a child element to a parent element
function appendChild(parent, child) {
  parent.appendChild(child);
}

// Function to remove all children of a parent element
function removeAllChildren(parent) {
  // Check if parent element exists
  if (!parent) {
    console.error('Parent element is not valid or does not exist.');
    return;
  }

  // Remove all children until there are none left
  while (parent.firstChild) {
    parent.removeChild(parent.firstChild);
  }
}

// EVENT HANDLING
// Function to handle key press events
function onEnterKeyPress(element, callback) {
  element.addEventListener("keydown", function (e) {
    if (e.key === "Enter") {
      callback();
    }
  });
}

// Function to handle click events
function onClick(element, callback) {
  element.addEventListener("click", callback);
}

