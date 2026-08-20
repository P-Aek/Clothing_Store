erDiagram

    USER ||--|| CART : has
    USER ||--o{ ORDER : places

    CART ||--o{ CART_ITEM : contains
    PRODUCT ||--o{ CART_ITEM : referenced_by

    ORDER ||--|{ ORDER_ITEM : contains

    PRODUCT ||--|{ VARIANT : has

    USER {
        ObjectId _id PK
        string name
        string email
        string passwordHash
        string role
        datetime createdAt
        datetime updatedAt
    }

    PRODUCT {
        ObjectId _id PK
        string name
        string description
        number price
        boolean active
        datetime createdAt
        datetime updatedAt
    }

    VARIANT {
        ObjectId _id
        string color
        string size
        number stock
    }

    CART {
        ObjectId _id PK
        ObjectId userId FK
        datetime createdAt
        datetime updatedAt
    }

    CART_ITEM {
        ObjectId productId FK
        ObjectId variantId
        number quantity
    }

    ORDER {
        ObjectId _id PK
        ObjectId userId FK
        number totalPrice
        string status
        datetime createdAt
        datetime updatedAt
    }

    ORDER_ITEM {
        ObjectId productId
        ObjectId variantId
        string productName
        string color
        string size
        number price
        number quantity
        number subtotal
    }