#  Command 
# 1. Quick Restart (Fastest - 3 seconds)
# bash
# docker compose restart backend
# docker compose logs -f backend
# Stops + starts backend only. DB stays untouched.

# 2. Full Restart (10 seconds)
# bash
# docker compose up -d backend
# docker compose logs -f backend

# ============================== #