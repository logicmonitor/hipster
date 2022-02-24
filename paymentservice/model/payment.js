const mongoose = require("mongoose");

const PaymentSchema = mongoose.Schema({
    cardNumber: {
        type: String,
        required: true
    },
    cardType: {
        type: String,
        required: true
    },
    amount: {
        type: String,
        required: true
    },
    createdAt: {
        type: Date,
        default: Date.now()
    }
});

module.exports = mongoose.model("payment", PaymentSchema);