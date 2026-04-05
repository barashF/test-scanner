import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
	stages: [
		{ duration: "5s", target: 20 },
		{ duration: "20s", target: 100 },
		{ duration: "5s", target: 0 },
	],
	thresholds: {
		http_req_failed: ["rate<0.01"],
		http_req_duration: ["p(95)<200"],
	},
};

export function setup() {
	const baseUrl = "http://localhost:8080";

	const loginPayload = JSON.stringify({ role: "user" });
	const loginParams = {
		headers: { "Content-Type": "application/json" },
	};

	const loginRes = http.post(
		`${baseUrl}/dummyLogin`,
		loginPayload,
		loginParams,
	);

	const checkLogin = check(loginRes, {
		"dummyLogin status is 200": (r) => r.status === 200,
	});

	if (!checkLogin) {
		throw new Error(
			`failed to get jwt token. status: ${loginRes.status}, body: ${loginRes.body}`,
		);
	}

	const loginData = JSON.parse(loginRes.body);
	const token = loginData.access_token;
	console.log("jwt token obtained successfully");

	const authParams = {
		headers: {
			Authorization: `Bearer ${token}`,
			"Content-Type": "application/json",
		},
	};

	const roomsRes = http.get(`${baseUrl}/rooms/list`, authParams);

	const checkRooms = check(roomsRes, {
		"rooms list status is 200": (r) => r.status === 200,
	});

	if (!checkRooms) {
		throw new Error(
			`failed to fetch rooms list. status: ${roomsRes.status}, body: ${roomsRes.body}`,
		);
	}

	const roomsData = JSON.parse(roomsRes.body);

	if (!roomsData.rooms || roomsData.rooms.length === 0) {
		throw new Error("rooms list is empty, run mock seeder first");
	}

	const targetRoomId = roomsData.rooms[0].id;

	return {
		roomId: targetRoomId,
		token: token,
	};
}

export default function (data) {
	const today = new Date().toISOString().slice(0, 10);

	const url = `http://localhost:8080/rooms/${data.roomId}/slots/list?date=${today}`;

	const params = {
		headers: {
			Authorization: `Bearer ${data.token}`,
			"Content-Type": "application/json",
		},
	};

	const res = http.get(url, params);

	check(res, {
		"status is 200": (r) => r.status === 200,
	});

	sleep(0.1);
}
