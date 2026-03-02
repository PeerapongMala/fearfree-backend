const fetchAPI = async (path: string, options: RequestInit = {}) => {
  const url = `http://localhost:8080/api/v1${path}`;
  const res = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options.headers,
    },
  });

  const text = await res.text();
  try {
    return { status: res.status, ok: res.ok, data: JSON.parse(text) };
  } catch (e) {
    return { status: res.status, ok: res.ok, data: text };
  }
};

const runTests = async () => {
  console.log("🚀 Starting API Tests...");
  const username = `testuser_${Date.now()}`;
  const password = "password123";

  // 1. Signup
  console.log("\n[1] Testing Signup...");
  let res = await fetchAPI("/auth/signup", {
    method: "POST",
    body: JSON.stringify({
      username,
      password,
      email: `${username}@example.com`,
      full_name: "Test User",
      age: 25,
      most_fear_animal: "Snake",
      fear_level: "high",
    }),
  });
  console.log("Signup:", res.status, res.data);
  if (!res.ok) return;

  // 2. Login
  console.log("\n[2] Testing Login...");
  res = await fetchAPI("/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
  console.log("Login:", res.status, res.data.message);
  if (!res.ok) return;

  const token = res.data.token;
  const headers = { Authorization: `Bearer ${token}` };

  // 3. Stages - Categories
  console.log("\n[3] Testing Get Categories...");
  res = await fetchAPI("/stages/categories", { headers });
  console.log("Categories:", res.status, res.data.data?.length, "items");
  if (!res.ok || !res.data.data || res.data.data.length === 0) {
    console.log("❌ No categories found or error.", res.data);
    return;
  }
  const categoryId = res.data.data[0].id;

  // 4. Stages - Animals
  console.log(`\n[4] Testing Get Animals for Category ${categoryId}...`);
  res = await fetchAPI(`/stages/categories/${categoryId}/animals`, { headers });
  console.log("Animals:", res.status, res.data.data?.length, "items");
  if (!res.ok || !res.data.data || res.data.data.length === 0) {
    console.log("❌ No animals found.", res.data);
    // It's possible DB is not seeded! We'll just continue to see.
  } else {
    const animalId = res.data.data[0].id;

    // 5. Stages - Levels
    console.log(`\n[5] Testing Get Levels for Animal ${animalId}...`);
    res = await fetchAPI(`/stages/animals/${animalId}/levels`, { headers });
    console.log("Levels:", res.status, res.data.data?.length, "items");

    if (res.ok && res.data.data && res.data.data.length > 0) {
      const levelId = res.data.data[0].id;

      // 6. Submit Stage Result
      console.log(`\n[6] Testing Submit Stage Result for Level ${levelId}...`);
      res = await fetchAPI(`/stages/levels/${levelId}/results`, {
        method: "POST",
        headers,
        body: JSON.stringify({ answer: "pass", symptom_note: "Doing okay" }),
      });
      console.log("Submit Stage:", res.status, res.data);
    } else {
      console.log("❌ No levels to test.");
    }
  }

  // 7. Rewards
  console.log("\n[7] Testing Get Rewards...");
  res = await fetchAPI("/rewards/", { headers });
  console.log("Rewards List:", res.status, res.data.data?.length, "items");

  if (res.ok && res.data.data && res.data.data.length > 0) {
    const rewardId = res.data.data[0].id;

    // 8. Redeem Reward
    console.log(`\n[8] Testing Redeem Reward ID ${rewardId}...`);
    res = await fetchAPI(`/rewards/${rewardId}/redeem`, {
      method: "POST",
      headers,
    });
    console.log("Redeem Reward:", res.status, res.data);

    // 9. My Redemptions
    console.log("\n[9] Testing My Redemptions...");
    res = await fetchAPI("/users/me/redemptions", { headers });
    console.log("My Redemptions:", res.status, res.data.data?.length, "items");
  } else {
    console.log("❌ No rewards to test redemption.");
  }

  console.log("\n🎉 API Tests Completed.");
};

runTests();
