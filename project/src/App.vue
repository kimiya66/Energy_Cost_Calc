<template>
  <div id="app">
    <div class="btns">
      <button @click="generateYesterdaysReadings()">Generate readings</button>
      <button @click="calculateCost()">Calculate cost</button>
    </div>
    <!-- show error message when the meter reading is empty-->
    <div v-if="errorMessage" class="error">{{ errorMessage }}</div>

    <div>
      <table>
      <thead>
        <tr>
          <th>Timestamp</th>
          <th>Energy Consumption (kwh)</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(reading, index) in meterReadings" :key="index">
          <td>{{ new Date(reading.timestamp).toLocaleString() }}</td>
          <td>{{ reading.reading.toFixed(2) }}</td>
        </tr>
      </tbody>
    </table>
    </div>
    <div>
      Cost: €{{ cost.toLocaleString('en', { maximumFractionDigits: 2, minimumFractionDigits: 2 }) }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, type Ref } from 'vue'
import axios from 'axios'

interface MeterReading {
  timestamp: Date
  reading: number
}

const meterReadings: Ref<MeterReading[]> = ref([])
const cost: Ref<number> = ref(0)
const errorMessage = ref('');

// Generates random meter readings spanning the 24 hours of yesterday's date.
// One meter reading per hour is generated and the readings are sorted ascending by time.
function generateYesterdaysReadings() {
  var kwh = Math.random() * 1000000
  const readings = []

  for (var hour = 0; hour <= 24; hour++) {
    const date = new Date()
    date.setDate(date.getDate() - 1)
    date.setHours(hour, 0, 0, 0)

    kwh += Math.random() * 100

    readings.push({
      timestamp: date,
      reading: kwh
    })
  }

  meterReadings.value = readings
}
// Calculate cost by sending the readings to the backend API
async function calculateCost() {
  if (!meterReadings.value || 
    meterReadings.value.length === 0 || 
    !meterReadings.value.every(r => r.timestamp && r.reading !== undefined)) {
    errorMessage.value = "Please provide meter readings before calculation.";
    setTimeout(() => {
      errorMessage.value = '';
    }, 3000); // Remove after 3 seconds
    return;
  }
  try {
    const response = await axios.post('http://localhost:8080/energy_cost',
      meterReadings.value.map(r => ({
        timestamp: new Date(r.timestamp).getTime(), // convert ISO date to milliseconds
        reading: r.reading
      })),
    {
        headers: {
          'Content-Type': 'application/json', // Set Content-Type
        },
      }
  );
    cost.value = response.data.total_cost
    errorMessage.value = "" // clear previous error messages
  } catch (error) {
    errorMessage.value = "Failed to submit readings. Please try again.";
    alert('Error calculating cost. Check console for details.')

  }
}
</script>

<style lang="css" scoped>
#app {
  padding: 30px;
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.btns {
  display: flex;
  gap: 8px;
}
.error {
    color: red;
    font-weight: bold;
    margin-top: 10px;
}
table {
  border-collapse: collapse;
  width: 100%;
}

th, td {
  border: 1px solid #ccc;
  padding: 8px;
  text-align: left;
}

th {
  background-color: #f4f4f4;
}
</style>
