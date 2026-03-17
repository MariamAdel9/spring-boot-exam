package com.mariam.exam2.service;

import com.mariam.exam2.models.Product;
import com.mariam.exam2.repository.ProductRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

/**
 * ProductService — business logic layer for products.
 *
 * YOUR TASK (Part B):
 *   1. Add constructor injection for ProductRepository
 *   2. Implement all 5 methods below
 *
 * Follow the same pattern from the BookService in Day 2:
 *   - Constructor takes the repository as a parameter
 *   - Each method delegates to the repository
 */
@Service

public class ProductService {

    // TODO: Declare a private final ProductRepository field
    // TODO: Constructor that takes ProductRepository as parameter (constructor injection)

    private final ProductRepository productRepository;

    public ProductService(ProductRepository productRepository){
        this.productRepository=productRepository;
    }

    /**
     * Get all products.
     */
    public List<Product> getAllProducts() {
        // TODO: Delegate to repository

        return productRepository.findAll();
    }


    /**
     * Get a product by ID.
     * Returns Optional<Product> — empty if not found.
     */
    public Optional<Product> getProductById(Long id) {
        // TODO: Delegate to repository
        return productRepository.findById(id);
    }


    /**
     * Create a new product.
     * @return the saved product (with generated ID)
     */
    public Product createProduct(Product product) {
        // TODO: Delegate to repository
        return productRepository.save(product);
    }


    /**
     * Update an existing product.
     * Find the existing product by ID, update its fields, and save it.
     *
     * @return Optional containing the updated product, or empty if not found
     */
    public Optional<Product> updateProduct(Long id, Product updated) {
        // TODO: Find existing product by ID
        // TODO: If found, update its name, category, price, and quantity
        // TODO: Save and return the updated product
        // TODO: If not found, return Optional.empty()

        //check if there are already saved product with this id
        Optional<Product> exists = productRepository.findById(id);

        //if yes the save it and set its filed to the new given filed
        if (exists.isPresent()) {
            Product existing = exists.get();

            existing.setName(updated.getName());
            existing.setCategory(updated.getCategory());
            existing.setPrice(updated.getPrice());
            existing.setQuantity(updated.getQuantity());

            //save the product again with same id
            productRepository.save(existing);

            return Optional.of(existing);
        }

        return Optional.empty();
    }


    /**
     * Delete a product by ID.
     * @return true if deleted, false if not found
     */
    public boolean deleteProduct(Long id) {
        // TODO: Delegate to repository
        return productRepository.deleteById(id);
    }
}

