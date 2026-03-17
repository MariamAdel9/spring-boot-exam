package com.mariam.exam2.controller;

import com.mariam.exam2.models.Product;
import com.mariam.exam2.service.ProductService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

/**
 * ProductController — REST API endpoints for products.
 *
 * <p>YOUR TASK (Part B): 1. Add constructor injection for ProductService 2. Implement all 5
 * endpoints below 3. Use ResponseEntity to return proper HTTP status codes: - GET all → 200 OK -
 * GET by id → 200 OK or 404 Not Found - POST create → 201 Created - PUT update → 200 OK or 404 Not
 * Found - DELETE → 204 No Content or 404 Not Found
 *
 * <p>Refer to the BookController from Day 2 for the pattern.
 */
@RestController
@RequestMapping("/api/products")
public class ProductController {


  // TODO: Declare a private final ProductService field
  // TODO: Constructor that takes ProductService as parameter
  private final ProductService productService;

  public ProductController(ProductService productService) {
    this.productService = productService;
  }
  /** GET /api/products Return all products with status 200. */
  @GetMapping
  public ResponseEntity<List<Product>> getAllProducts() {
    return ResponseEntity.ok(productService.getAllProducts());

  }

  /** GET /api/products/{id} Return the product with status 200, or 404 if not found. */
  @GetMapping("/{id}")
  public ResponseEntity<Product> getProductById(@PathVariable Long id) {
    return productService.getProductById(id).map(ResponseEntity::ok).orElse(ResponseEntity.notFound().build());
  }

  /** POST /api/products Create a new product. Return it with status 201 Created. */
  @PostMapping
  public ResponseEntity<Product> createProduct(@RequestBody Product product) {
    // TODO: Implement
    // Hint: use ResponseEntity.status(HttpStatus.CREATED).body(...)
    return ResponseEntity.status(HttpStatus.CREATED).body( productService.createProduct(product));
  }

  /**
   * PUT /api/products/{id} Update an existing product. Return 200 with updated product, or 404 if
   * not found.
   */
  @PutMapping("/{id}")
  public ResponseEntity<Product> updateProduct(
      @PathVariable Long id, @RequestBody Product product) {
    // TODO: Implement
    return productService.updateProduct(id,product).map(ResponseEntity::ok).orElse(ResponseEntity.notFound().build());

  }

  /**
   * DELETE /api/products/{id} Delete a product. Return 204 No Content if deleted, 404 if not found.
   */
  @DeleteMapping("/{id}")
  public ResponseEntity<Void> deleteProduct(@PathVariable Long id) {
    // TODO: Implement
    // Hint: return ResponseEntity.noContent().build() for success
    return productService.deleteProduct(id)? ResponseEntity.noContent().build():ResponseEntity.notFound().build();
  }
}
