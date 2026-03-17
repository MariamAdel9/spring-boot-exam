Title and description:
product manager is A simple REST API built with Spring Boot to 
manage products (create, read, update, delete) with a built-in memory. 

How to run:
to run open a terminal and write the following command (mvnw spring-boot:run)
and the project will start running with tomcat in port 3000 (http://localhost:3000)

API Endpoint:
the program has the following api end points 

Method  	URL	                Description

GET	        /api/products	    Get all products
GET	        /api/products/{id}	Get product by ID via path variable
POST	    /api/products	    Create a new product
PUT	        /api/products/{id}	Update product
DELETE	    /api/products/{id}	Delete product

Crul Example (I used postman to test my api points):

**1- To create a new product** 

curl --location 'http://localhost:3000/api/products' \
--header 'Content-Type: application/json' \
--data '{
"name": "Laptop",
"category": "Electronics",
"price": 999.99,
"quantity": 10
}'

**2-To get all products**
curl -X GET http://localhost:3000/api/products

**3- To get a product by its id**
curl -X GET http://localhost:3000/api/products/1

**4- To update a product**
curl --location --request PUT 'http://localhost:3000/api/products/2' \
--header 'Content-Type: application/json' \
--data '{
"name": "Harry poter book",
"category": "book",
"price": 5.99,
"quantity": 10
}'

Sample response:

**#save a product:**

1-valid body
{
"name": "Water",
"category": "Food",
"price": 1.99,
"quantity": 10
}

result (status code 201)
{
"name": "Water",
"category": "Food",
"price": 1.99,
"quantity": 10,
"id": 4
}


2-missing body 
{
"name": "Laptop",
"category": "Electronics",
"price": 999.99
}

result (status code 400)
{
"timestamp": "2026-03-17T15:57:03.047Z",
"status": 400,
"error": "Bad Request",
"trace": "org.Caused...",
"message": "JSON parse error: Cannot map `null` into type `int`...",
"path": "/api/products"
}

**#get all products:**

result:
[
{
"name": "Laptop",
"category": "Electronics",
"price": 999.99,
"quantity": 10,
"id": 1
},
{
"name": "Laptop",
"category": "Electronics",
"price": 999.99,
"quantity": 10,
"id": 2
},
{
"name": "Laptop",
"category": "",
"price": 999.99,
"quantity": 10,
"id": 3
},
{
"name": "Water",
"category": "Food",
"price": 1.99,
"quantity": 10,
"id": 4
}
]

**#get a product by id**

1-search for a product with a valid id
http://localhost:3000/api/products/1 

result: 
{
"name": "Laptop",
"category": "Electronics",
"price": 999.99,
"quantity": 10,
"id": 1
}

2-search for a product with a non-valid id

result:
404 Not Found

**#update a product**

1-update product with existing id
http://localhost:3000/api/products/2
Body
{
"name": "Harry poter book",
"category": "book",
"price": 5.99,
"quantity": 10
}

result:
{
"name": "Harry poter book",
"category": "book",
"price": 5.99,
"quantity": 10,
"id": 2
}

2-update product with non-existing id

http://localhost:3000/api/products/10

result:
404 Not Found

**#delete a product**

1-delete a product with an existing an id 

http://localhost:3000/api/products/1

result:
204 No Content


**#delete a product**

1-delete a product with a non-existing an id

http://localhost:3000/api/products/6

result:
404 Not Found

**#final array list after all requests**

[
{
"name": "Harry poter book",
"category": "book",
"price": 5.99,
"quantity": 10,
"id": 2
},
{
"name": "Laptop",
"category": "",
"price": 999.99,
"quantity": 10,
"id": 3
},
{
"name": "Water",
"category": "Food",
"price": 1.99,
"quantity": 10,
"id": 4
}
]