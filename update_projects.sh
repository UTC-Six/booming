# ... existing script ...

# 添加初始化数据库索引的步骤
mysql -u username -p'your_password' your_database <<EOF
CREATE INDEX idx_landmines_lat_lon ON landmines (latitude, longitude);
ALTER TABLE landmines ADD location POINT NOT NULL;
UPDATE landmines SET location = POINT(longitude, latitude);
CREATE SPATIAL INDEX idx_landmines_location ON landmines(location);
EOF

# ... existing script ... 