// static/manage-vocabulary.js

// Function to fetch dataset items from the server
async function fetchDatasetItems() {
    try {
        const response = await fetch("http://localhost:8080/dataset");
        const data = await response.json();
        return data;
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
        questionElement.textContent = item.question;
        itemElement.appendChild(questionElement);

        const categoryElement = document.createElement("p");
        categoryElement.textContent = "Category: " + item.category;
        itemElement.appendChild(categoryElement);

        const updateButton = document.createElement("button");
        updateButton.textContent = "Update";
        updateButton.addEventListener("click", () =>  updateDatasetItem(item.id)); 
        itemElement.appendChild(updateButton);

        const deleteButton = document.createElement("button");
        deleteButton.textContent = "Delete";
        deleteButton.addEventListener("click", () => deleteDatasetItem(item.id));
        itemElement.appendChild(deleteButton);

        datasetContainer.appendChild(itemElement);
    });
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
        for (let key in currentItem.answers) {
            const answerLabel = document.createElement("label");
            answerLabel.textContent = `${key}:`;
            updateForm.appendChild(answerLabel);

            const answerInput = document.createElement("input");
            answerInput.type = "text";
            answerInput.value = currentItem.answers[key];
            answerInput.name = key;
            updateForm.appendChild(answerInput);
        }

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
                answers: {},
                correct: parseInt(correctInput.value)
            };

            // Assign values from answer inputs to formData.answers object
            for (let key in currentItem.answers) {
                formData.answers[key] = updateForm.elements[key].value;
            }

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

    const formData = {
        category: document.getElementById("category").value,
        question: document.getElementById("question").value,
        targetWord: document.getElementById("targetWord").value,
        answers: {
            [document.getElementById("answer1Text").value]: document.getElementById("answer1Image").value,
            [document.getElementById("answer2Text").value]: document.getElementById("answer2Image").value,
            [document.getElementById("answer3Text").value]: document.getElementById("answer3Image").value,
            [document.getElementById("answer4Text").value]: document.getElementById("answer4Image").value
        },
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
