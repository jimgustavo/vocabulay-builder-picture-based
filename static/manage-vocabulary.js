// static/manage-vocabulary.js

// Function to fetch dataset items from the server
async function fetchDatasetItems() {
    try {
        const response = await fetch("http://localhost:8080/dataset");
        const data = await response.json();
        return data.map(item => ({
            id: item.id,
            category: item.category,
            question: item.question,
            targetWord: item.targetWord,
            answers: Object.keys(item.answers).map(option => ({
                //option: option,
                //url: item.answers[option]
                option: item.answers[option], // Use item.answers[option] as the option value
                url: option // Use option as the URL value
            })),
            correct: item.correct
        }));
    } catch (error) {
        console.error("Error fetching dataset items:", error);
        return [];
    }
}

// Function to display dataset items
async function displayDatasetItems() {
    const datasetItems = await fetchDatasetItems();
    const datasetContainer = document.getElementById("datasetContainer");
    datasetContainer.innerHTML = ""; // Clear previous items

    datasetItems.forEach(item => {
        const itemElement = document.createElement("div");
        itemElement.classList.add("dataset-item");

        const questionElement = document.createElement("p");
        questionElement.textContent = "Question: " + item.question;
        itemElement.appendChild(questionElement);

        const categoryElement = document.createElement("p");
        categoryElement.textContent = "Category: " + item.category;
        itemElement.appendChild(categoryElement);

        // Create a <ul> element
        const answersList = document.createElement("ul");

        // Iterate over each answer option in the item and create <li> elements
        item.answers.forEach(answer => {
            const answerItem = document.createElement("li");
            answerItem.textContent = `${answer.url}: ${answer.option.option}`; // Access the url property of the answer object
            answersList.appendChild(answerItem);
        });

        // Append the <ul> element to the item element
        itemElement.appendChild(answersList);

        const correctElement = document.createElement("p");
        correctElement.textContent = "Correct: " + item.correct;
        itemElement.appendChild(correctElement);
        
        const updateButton = document.createElement("button");
        updateButton.textContent = "Update";
        updateButton.addEventListener("click", () =>  updateDatasetItem(item.id)); 
        itemElement.appendChild(updateButton);

        const deleteButton = document.createElement("button");
        deleteButton.textContent = "Delete";
        deleteButton.addEventListener("click", () => deleteDatasetItem(item.id));
        itemElement.appendChild(deleteButton);

        const duplicateButton = document.createElement("button");
        duplicateButton.textContent = "Duplicate";
        duplicateButton.addEventListener("click", () => duplicateDatasetItem(item.id));
        itemElement.appendChild(duplicateButton);

        const scrambleButton = document.createElement("button");
        scrambleButton.textContent = "Scramble Answers";
        scrambleButton.addEventListener("click", () => scrambleAnswers(item.id)); 
        itemElement.appendChild(scrambleButton);

        datasetContainer.appendChild(itemElement);
    });

     // Populate category filter dropdown
     const categories = [...new Set(datasetItems.map(item => item.category))];
     const categoryFilterDropdown = document.getElementById("categoryFilter");
     categoryFilterDropdown.innerHTML = "<option value=''>All</option>";
     categories.forEach(category => {
         const option = document.createElement("option");
         option.value = category;
         option.textContent = category;
         categoryFilterDropdown.appendChild(option);
     });
}

// Function to filter dataset items by category
function filterByCategory() {
    const categoryFilter = document.getElementById("categoryFilter").value;
    const datasetItems = document.querySelectorAll(".dataset-item");

    datasetItems.forEach(item => {
        if (categoryFilter === "" || item.querySelector("p:nth-child(2)").textContent.includes(categoryFilter)) {
            item.style.display = "block";
        } else {
            item.style.display = "none";
        }
    });
}

async function scrambleAnswers(itemId) {
    try {
        const response = await fetch(`http://localhost:8080/dataset/${itemId}/scramble`, {
            method: "POST"
        });
        if (response.ok) {
            console.log(`Answers for dataset item with ID ${itemId} scrambled successfully.`);
            // Update the displayed dataset items after successful scrambling
            displayDatasetItems();
        } else {
            const responseData = await response.text();
            console.log("Error scrambling answers:", responseData);
        }
    } catch (error) {
        console.error("Error:", error);
    }
}

// Function to duplicate a dataset item
async function duplicateDatasetItem(id) {
    try {
        const response = await fetch(`http://localhost:8080/dataset/${id}/duplicate`, {
            method: "POST"
        });
        if (response.ok) {
            console.log(`Dataset item with ID ${id} duplicated successfully.`);
            // Update the displayed dataset items after successful duplication
            displayDatasetItems();
        } else {
            const responseData = await response.text();
            console.log("Error duplicating dataset item:", responseData);
        }
    } catch (error) {
        console.error("Error:", error);
    }
}

// Function to update a dataset item
async function updateDatasetItem(id) {
    try {
        // Retrieve current data of the dataset item with the specified ID
        const response = await fetch(`http://localhost:8080/dataset/${id}`);
        const currentItem = await response.json();

        // Display update form below the displayed items
        const updateForm = document.createElement("form");
        updateForm.id = "updateForm";

        // Add title to the form
        const title = document.createElement("h1");
        title.textContent = "Update Dataset";
        updateForm.appendChild(title);

        // Category
        const categoryLabel = document.createElement("label");
        categoryLabel.textContent = "Category:";
        updateForm.appendChild(categoryLabel);

        const categoryInput = document.createElement("input");
        categoryInput.type = "text";
        categoryInput.value = currentItem.category;
        updateForm.appendChild(categoryInput);

        // Question
        const questionLabel = document.createElement("label");
        questionLabel.textContent = "Question:";
        updateForm.appendChild(questionLabel);

        const questionInput = document.createElement("input");
        questionInput.type = "text";
        questionInput.value = currentItem.question;
        updateForm.appendChild(questionInput);

        // Target Word
        const targetWordLabel = document.createElement("label");
        targetWordLabel.textContent = "Target Word:";
        updateForm.appendChild(targetWordLabel);

        const targetWordInput = document.createElement("input");
        targetWordInput.type = "text";
        targetWordInput.value = currentItem.targetWord;
        updateForm.appendChild(targetWordInput);

        // Answers
        currentItem.answers.forEach(answer => {
            const answerLabel = document.createElement("label");
            answerLabel.textContent = `${answer.option}:`;
            updateForm.appendChild(answerLabel);

            const answerInput = document.createElement("input");
            answerInput.type = "text";
            answerInput.value = answer.url;
            answerInput.name = answer.option;
            updateForm.appendChild(answerInput);
        });

        // Correct Answer
        const correctLabel = document.createElement("label");
        correctLabel.textContent = "Correct Answer:";
        updateForm.appendChild(correctLabel);

        const correctInput = document.createElement("input");
        correctInput.type = "number";
        correctInput.value = currentItem.correct;
        updateForm.appendChild(correctInput);

        const submitButton = document.createElement("button");
        submitButton.type = "submit";
        submitButton.textContent = "Update";
        updateForm.appendChild(submitButton);

        // Handle form submission
        updateForm.addEventListener("submit", async function(event) {
            event.preventDefault();

            const formData = {
                category: categoryInput.value,
                question: questionInput.value,
                targetWord: targetWordInput.value,
                answers: [],
                correct: parseInt(correctInput.value)
            };

            // Assign values from answer inputs to formData.answers array
            currentItem.answers.forEach(answer => {
                formData.answers.push({
                    option: answer.option,
                    url: updateForm.elements[answer.option].value
                });
            });

            try {
                const updateResponse = await fetch(`http://localhost:8080/dataset/${id}`, {
                    method: "PUT",
                    headers: {
                        "Content-Type": "application/json"
                    },
                    body: JSON.stringify(formData)
                });

                if (updateResponse.ok) {
                    console.log(`Dataset item with ID ${id} updated successfully.`);
                    // Update the displayed dataset items after successful update
                    displayDatasetItems();
                } else {
                    const responseData = await updateResponse.text();
                    console.log("Error updating dataset item:", responseData);
                }
            } catch (error) {
                console.error("Error:", error);
            }
        });

        // Append the update form below the displayed items
        const datasetContainer = document.getElementById("datasetContainer");
        datasetContainer.appendChild(updateForm);

    } catch (error) {
        console.error("Error:", error);
    }
}

// Function to delete a dataset item
async function deleteDatasetItem(id) {
    try {
        const response = await fetch(`http://localhost:8080/dataset/${id}`, {
            method: "DELETE"
        });
        if (response.ok) {
            console.log(`Dataset item with ID ${id} deleted successfully.`);
            // Call displayDatasetItems() after deleting an item
            displayDatasetItems();
        } else {
            const responseData = await response.text();
            console.log("Error deleting dataset item:", responseData);
        }
    } catch (error) {
        console.error("Error:", error);
    }
}

// Call displayDatasetItems() to initially display dataset items
displayDatasetItems();

document.getElementById("datasetForm").addEventListener("submit", async function(event) {
    event.preventDefault();

    const answers = [
        {
            option: document.getElementById("answer1Text").value,
            url: document.getElementById("answer1Image").value
        },
        {
            option: document.getElementById("answer2Text").value,
            url: document.getElementById("answer2Image").value
        },
        {
            option: document.getElementById("answer3Text").value,
            url: document.getElementById("answer3Image").value
        },
        {
            option: document.getElementById("answer4Text").value,
            url: document.getElementById("answer4Image").value
        }
    ];

    const formData = {
        category: document.getElementById("category").value,
        question: document.getElementById("question").value,
        targetWord: document.getElementById("targetWord").value,
        answers: answers,
        correct: parseInt(document.getElementById("correct").value)
    };

    // Log the JSON data being sent in the request
    console.log("JSON Data Sent:", JSON.stringify(formData));

    try {
        const response = await fetch("http://localhost:8080/dataset", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify(formData)
        });
        if (response.ok) {
            document.getElementById("message").innerText = "Dataset item added successfully.";
            // Clear form fields after successful submission
            document.getElementById("datasetForm").reset();
            // Call displayDatasetItems() after adding a new item
            displayDatasetItems();
        } else {
            const responseData = await response.text();
            console.log("Response data:", responseData);
            document.getElementById("message").innerText = "Error: " + responseData;
        }
    } catch (error) {
        console.error("Error:", error);
        document.getElementById("message").innerText = "Error: Failed to add dataset item.";
    }
});
