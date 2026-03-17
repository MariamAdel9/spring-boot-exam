package com.mariam.exam2.models;

import lombok.ToString;

/**
 * Product — represents a product in the inventory.
 *
 * <p>YOUR TASK (Part A): 1. Add a no-arg constructor (required by Spring/Jackson for JSON
 * deserialization) 2. Add a constructor that takes all fields EXCEPT id: Product(String name,
 * String category, double price, int quantity) 3. Add getters and setters for ALL fields (including
 * id) 4. Override toString() to return a readable string like: "Product{id=1, name='Laptop',
 * category='Electronics', price=999.99, quantity=10}"
 */
@ToString
public class Product {

  private Long id;
  private String name;
  private String category;
  private double price;
  private int quantity;

  // TODO: No-arg constructor

  public Product() {}

  // TODO: Constructor with (String name, String category, double price, int quantity)

  public Product(String name, String category, double price, int quantity) {
    this.name = name;
    this.category = category;
    this.price = price;
    this.quantity = quantity;
  }

  public Long getId() {
    return id;
  }

  public void setId(Long id) {
    this.id = id;
  }

  public String getName() {
    return name;
  }

  public void setName(String name) {
    this.name = name;
  }

  public String getCategory() {
    return category;
  }

  public void setCategory(String category) {
    this.category = category;
  }

  public double getPrice() {
    return price;
  }

  public void setPrice(double price) {
    this.price = price;
  }

  public int getQuantity() {
    return quantity;
  }

  public void setQuantity(int quantity) {
    this.quantity = quantity;
  }

  // TODO: Override toString()
  @Override
  public String toString() {
      return "Product{id=" + getId() + ", name='" + getName() + ", category='" + getCategory() +", price=" + getPrice() +", quantity=" + getQuantity() +'}';
  }

}
