import { check, sleep } from 'k6';
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '10s', target: 1 },
    { duration: '30s', target: 5 },
    { duration: '20s', target: 0 },
  ],
  thresholds: {
    'http_req_duration': ['p(95)<300'],
    'http_req_failed': ['rate<0.001'],
  },
};

export function setup() {
  const baseUrl = 'http://app:8080';
  
  const teamPayload = JSON.stringify({
    team_name: 'loadtest-team',
    members: [
      { user_id: 'lt1', username: 'loaduser1', is_active: true },
      { user_id: 'lt2', username: 'loaduser2', is_active: true },
      { user_id: 'lt3', username: 'loaduser3', is_active: true },
    ]
  });
  
  try {
    http.post(`${baseUrl}/team/add`, teamPayload, {
      headers: { 'Content-Type': 'application/json' },
    });
  } catch (e) {
  }
  
  return { 
    baseUrl,
    teamName: 'loadtest-team',
    userIds: ['lt1', 'lt2', 'lt3']
  };
}

export default function (data) {
  const { baseUrl, teamName, userIds } = data;
  
  const endpoints = [
    { url: `${baseUrl}/health`, name: 'Health' },
    { url: `${baseUrl}/team/get?team_name=${teamName}`, name: 'Team' },
    { url: `${baseUrl}/users/getReview?user_id=${userIds[0]}`, name: 'User' },
    { url: `${baseUrl}/stats/review-assignments`, name: 'Stats' },
  ];
  
  const endpoint = endpoints[Math.floor(Math.random() * endpoints.length)];
  
  const response = http.get(endpoint.url);
  
  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 1s': (r) => r.timings.duration < 1000,
    'has valid response': (r) => r.body && r.body.length > 0,
  });
  
  if (response.status !== 200) {
    console.log(`❌ ${endpoint.url}: ${response.status} - ${response.body}`);
  }
  
  sleep(0.2);
}