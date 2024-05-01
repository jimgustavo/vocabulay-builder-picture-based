// static/script.js
var quiz; // Declare quiz variable in a scope accessible to event handler
var mastered = [];  // Set the mastered array to cero 
var learning = [];  // Set the learning array to cero 
var drillingMastered = 1; // Set the mastered words loop to the desired number 
var drillingLearning = 2; // Set the learning words loop to the desired number
var vocabularyStages = {
  "easy-stage": "http://localhost:8080/dataset/category/easy-word",
  "hard-stage": "http://localhost:8080/dataset/category/hard-word",
};

// Function to fetch dataSet from the API based on stage
async function fetchData(stage) {
  try {
    const response = await fetch(vocabularyStages[stage]);
    const data = await response.json();
    return data.map(item => ({
      id: item.id,
      question: item.question,
      targetWord: item.targetWord,
      correct: item.correct,
      answerImages: Object.values(item.answers),
      answers: Object.keys(item.answers)
    }));
  } catch (error) {
    console.error('Error fetching dataSet:', error);
    return [];
  }
}

// Function to shuffle the dataSet array
function shuffleArray(array) {
  for (var i = array.length - 1; i > 0; i--) {
    var j = Math.floor(Math.random() * (i + 1));
    var temp = array[i];
    array[i] = array[j];
    array[j] = temp;
  }
  return array;
}

// Function to initialize the vocabulary builder with stages
async function initializeVocabularyBuilder() {
  var currentStage = 0; // Index of current stage
  var stages = Object.keys(vocabularyStages); // Array of stage keys

  async function getNextStageData() {
    if (currentStage < stages.length) {
      var stageKey = stages[currentStage];
      var dataSet = await fetchData(stageKey);
      currentStage++;
      return dataSet;
    } else {
      return null; // No more stages
    }
  }

  async function startNextStage() {
    var dataSet = await getNextStageData();
    if (dataSet !== null) {
      mastered = [];
      // Continue with the new dataSet
      quiz = await vocabularyBuilder(dataSet); // Await the Promise here
      //quiz.checkAnswer(0); // Start the quiz immediately
    } else {
      // No more stages, display completion message
      alert("You have mastered all the vocabulary!");
    }
  }
  
  async function vocabularyBuilder(dataSet) {
    console.log("Building quiz with dataSet:", dataSet);
    var shuffledDataSet = shuffleArray(dataSet.slice()); // Make a copy of the original dataSet to shuffle

    var currentIndex = 0;
    var masteredIndex = 0;
    var learningIndex = 0;
    var masteredProgress = 0;
    var learningProgress = 0;

    // Update the displayQuestion function to remove text content setting for answer buttons

// Update the displayQuestion function to correctly display the answers with images
function displayQuestion() {
  var currentQuestion = shuffledDataSet[currentIndex];
  var questionText = currentQuestion.question;
  var targetWord = currentQuestion.targetWord;

  // Format the question string to make the target word bold
  var formattedQuestion = questionText.replace(targetWord, '<b>' + targetWord + '</b>');

  var questionTextElement = document.getElementById("question-text");
  questionTextElement.innerHTML = formattedQuestion;

  // Update progress counter
  var progressCounterElement = document.getElementById("progress-counter");
  progressCounterElement.textContent = (currentIndex + 1) + "/" + shuffledDataSet.length;

  // Update mastered counter
  var masteredCounterElement = document.getElementById("mastered-counter");
  masteredCounterElement.textContent = mastered.length + "/" + shuffledDataSet.length;

  // Update learning counter
  var learningCounterElement = document.getElementById("learning-counter");
  learningCounterElement.textContent = learning.length + "/" + shuffledDataSet.length;

  // Update progress bar
  var progressBarElement = document.getElementById("progress-bar");
  var progress = ((currentIndex + 1) / shuffledDataSet.length) * 100; // Calculate progress percentage
  progressBarElement.style.width = progress + "%";

  // Update mastered bar
  var masteredBarElement = document.getElementById("mastered-bar");
  masteredBarElement.style.width = masteredProgress + "%";

  // Update learning bar
  var learningBarElement = document.getElementById("learning-bar");
  learningBarElement.style.width = learningProgress + "%";

  // Display answers with images
  var answerElements = document.getElementsByClassName("answer-button");
  for (var i = 0; i < answerElements.length; i++) {
    var imgElement = answerElements[i].querySelector("img");
    imgElement.src = currentQuestion.answerImages[i]; // Set the src attribute of the image
    imgElement.alt = currentQuestion.answers[i]; // Set the alt attribute of the image
  }
}

    function checkAnswer(selectedIndex) {
      var currentQuestion = shuffledDataSet[currentIndex];
      if (selectedIndex === currentQuestion.correct) {
        masteredProgress = ((1 + masteredIndex++) * 1 / shuffledDataSet.length) * 100;
        // Repeat pushing the current question into the mastered array
        for (let i = 0; i < drillingMastered; i++) {
          mastered.push(currentQuestion);
        }
      } else {
        // Repeat pushing the current question into the learning array
        learningProgress = ((1 + learningIndex++) * 1 / shuffledDataSet.length) * 100;
        for (let i = 0; i < drillingLearning; i++) {
          learning.push(currentQuestion);
        }
      }

      currentIndex++;
      if (currentIndex < shuffledDataSet.length) {
        displayQuestion();
      } else if (currentIndex === shuffledDataSet.length && learning.length > 0) {
        shuffledDataSet = shuffleArray([...mastered, ...learning]);
        currentIndex = 0;
        masteredIndex = 0;
        learningIndex = 0;
        mastered = [];
        learning = [];
        masteredProgress = 0;
        learningProgress = 0;
        displayQuestion();
      } else if (currentIndex === shuffledDataSet.length && learning.length === 0) {
        startNextStage(); // Move to the next stage
      } else {
        console.log("something went wrong!");
      }
       // Print the updated array with elements ordered to be presented to the user
       console.log("Mastered:");
       console.log(mastered);
       console.log("Learning:");
       console.log(learning);
       console.log("Shuffled Data Set:");
       console.log(shuffledDataSet);

    }

    displayQuestion();

    return {
      checkAnswer: checkAnswer
    };
  }

  // Start with the first stage
  await startNextStage();
}

// Initialize the vocabulary builder
initializeVocabularyBuilder();


