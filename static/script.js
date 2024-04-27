// static/script.js

var quiz; // Declare quiz variable in a scope accessible to event handler
var mastered = [];
var learning = [];

// Function to fetch dataSet from the API
async function fetchData() {
  try {
    const response = await fetch('http://localhost:8080/dataset');
    const data = await response.json();
    return data;
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

// Function to initialize the vocabulary builder
async function initializeVocabularyBuilder() {

  // Fetch dataSet from the API
  var dataSet = await fetchData();

  function vocabularyBuilder() {
    var currentIndex = 0;
    var masteredIndex = 0;
    var learningIndex = 0;
    var masteredProgress = 0;
    var learningProgress = 0;
    var shuffledDataSet = shuffleArray(dataSet.slice()); // Make a copy of the original dataSet to shuffle

    function displayQuestion() {
      var currentQuestion = shuffledDataSet[currentIndex];
      var questionText = currentQuestion.question;
      var targetWord = currentQuestion.targetWord;
      var picturePath = currentQuestion.picture;

       // Format the question string to make the target word bold
      var formattedQuestion = questionText.replace(targetWord, '<b>' + targetWord + '</b>');

      var questionTextElement = document.getElementById("question-text");
      questionTextElement.innerHTML = formattedQuestion;

      var questionImageElement = document.getElementById("question-image");
      questionImageElement.src = picturePath;

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
      
      var answerElements = document.getElementsByName("answer");
      for (var i = 0; i < answerElements.length; i++) {
        answerElements[i].value = currentQuestion.answers[i]; // Set the value of the button
        answerElements[i].textContent = currentQuestion.answers[i]; // Set the text content of the button
      }
    }
    
    function checkAnswer(selectedIndex) {
      var currentQuestion = shuffledDataSet[currentIndex];
      if (selectedIndex === currentQuestion.correct) {
          //alert("Good job! The answer is correct!");
          //shuffledDataSet.splice(currentIndex, 1); // Remove the correctly answered question from the array
          mastered.push(currentQuestion);
          masteredProgress = (( 1 + masteredIndex++) * 1 / shuffledDataSet.length) * 100;
      } else {
          alert("The answer is incorrect!");
          //shuffledDataSet.splice(currentIndex, 1); // Remove the correctly answered question from the array
          learning.push(currentQuestion);
          learning.push(currentQuestion);
          learningProgress = (( 1 + learningIndex++) * 1 / shuffledDataSet.length) * 100;
      }

      // Move to the next question
      currentIndex++;
      if (currentIndex < shuffledDataSet.length) {
          displayQuestion();
      } else if (currentIndex === shuffledDataSet.length && learning.length > 0){
          //alert("End of questions.");
          shuffledDataSet = shuffleArray([...mastered, ...learning]);
          currentIndex = 0; // Reset currentIndex to 0 if all questions have been answered once
          masteredIndex = 0;
          learningIndex = 0;
          mastered = [];
          learning = [];
          masteredProgress = 0;
          learningProgress = 0;
          displayQuestion();
      } else if (currentIndex = shuffledDataSet.length && learning.length === 0) {
          alert("You have mastered all the vocabulary!");
          } else {
              console.log("something went wrong!")
          }

      // Print the updated array with elements ordered to be presented to the user
      //console.log("Mastered:");
      //console.log(mastered);
      //console.log("Learning:");
      //console.log(learning);
      //console.log("Shuffled Data Set:");
      //console.log(shuffledDataSet);
  }

    displayQuestion();

    return {
      checkAnswer: checkAnswer
    };
  }
  quiz = vocabularyBuilder(); // Assign quiz variable here
}

// Initialize the vocabulary builder
initializeVocabularyBuilder();

/*
var dataSet = [
  {
    id: 1,
    category: "easy-word",
    question: "This farm yielded very well this year.",
    targetWord: "yielded",
    picture: "https://images.pexels.com/photos/22192200/pexels-photo-22192200/free-photo-of-farm-yielded-very-well.jpeg?auto=compress&cs=tinysrgb&w=1260&h=750&dpr=1",
    answers: [
      "produced",
      "performed",
      "showed",
      "fell down"
    ],
    correct: 0
  },
  {
    id: 2,
    category: "hard-word",
    question: "Jannet didn't appreciate her boss's dig about her hairstyle.",
    targetWord: "dig",
    picture: "https://images.pexels.com/photos/22194116/pexels-photo-22194116.jpeg?auto=compress&cs=tinysrgb&w=1260&h=750&dpr=1",
    answers: [
      "critic",
      "compliment",
      "hole",
      "critical remark"
    ],
    correct: 3
  },
  {
    id: 3,
    category: "easy-word",
    question: "The assistant ushered the visitor to the boss's office",
    targetWord: "ushered",
    picture: "https://images.pexels.com/photos/22468285/pexels-photo-22468285.jpeg?auto=compress&cs=tinysrgb&w=1260&h=750&dpr=1",
    answers: [
      "asked",
      "showed",
      "walked someone to position",
      "assigned"
    ],
    correct: 2
  },
  {
    id: 4,
    category: "easy-word",
    question: "American officials are fearful of upending trade negotiations since they could be harmful ",
    targetWord: "upending",
    picture: "trade_negotiations.jpg",
    answers: [
      "v. improving",
      "v. denying",
      "v. cutting",
      "v. changing drastically"
    ],
    correct: 3
  },
  {
    id: 5,
    category: "easy-word",
    question: "Elizabeth always dithers for a while before she acts",
    targetWord: "dithers",
    picture: "elizabeth_dithers.jpg",
    answers: [
      "v. thinks",
      "v. hesitates",
      "v. forgets",
      "v. speaks"
    ],
    correct: 1
  },
  {
    id: 6,
    category: "hard-word",
    question: "Huawei, a telecoms giant, is in the blacklist since May over concerns that Chinese spooks use its gears to spy on America.",
    targetWord: "spooks",
    picture: "https://static.tvtropes.org/pmwiki/pub/images/origins_keyser_soze_the_unknown_foreigner_397373.png",
    answers: [
      "n. watcher",
      "n. worker",
      "n. spy",
      "n. soldier"
    ],
    correct: 2
  },
  {
    id: 7,
    category: "hard-word",
    question: "Huawei have been hoarding parts in anticipation of a ban and have sought other suppliers",
    targetWord: "hoarding",
    picture: "https://content.jdmagicbox.com/comp/indore/z7/0731px731.x731.140130184531.j8z7/catalogue/suvidha-automobiles-rnt-road-indore-automobile-part-dealers-maruti-1obcwk1w2j.jpg",
    answers: [
      "v. keeping for future",
      "v. wasting",
      "v. needing",
      "v. buying"
    ],
    correct: 0
  },
  {
    id: 8,
    category: "hard-word",
    question: "A shortfall in recruitment led to the company being understaffed  ",
    targetWord: "shortfall",
    picture: "shortfall_recruitment.jpg",
    answers: [
      "n. abundance",
      "n. deficit, less than needed",
      "n. ware",
      "n. team"
    ],
    correct: 1
  },
  {
    id: 9,
    category: "hard-word",
    question: "I think it's important not to downplay the significance of the event.",
    targetWord: "downplay",
    picture: "downplay_event.jpg",
    answers: [
      "v. play carefully",
      "v. enhance",
      "v. play hard",
      "v. minimize importance of"
    ],
    correct: 3
  }
];
*/