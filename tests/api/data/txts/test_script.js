// Test JavaScript file for SparkPlay upload testing
function greet(name) {
    return `Hello, ${name}! Welcome to SparkPlay.`;
}

const testData = {
    message: "This is a test file",
    timestamp: Date.now(),
    items: [1, 2, 3, 4, 5]
};

console.log(greet("Tester"));
console.log("Test data:", JSON.stringify(testData, null, 2));

module.exports = { greet, testData };
