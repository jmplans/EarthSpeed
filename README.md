# EarthSpeed

simple tool that calculates the rotational speed of any country on Earth
```
% go run . country
```

## Some maths

Because Earth is, what we call, an 'oblate spheroid', I wanted to move from sphere-like to ellipsoid calculation of the speed at a specific latitude/longitude.

The angular velocity of Earth (≈ 7.292 × 10⁻⁵ rad s⁻¹) is fixed; the variable part is the radius of the latitude circle the point sits on. On a perfect sphere that radius is simply:

```
R · cos φ
```

... because every meridian is a great circle.
On an ellipsoid that slice is not a circle of constant radius; the “horizontal” distance from the spin axis depends on the local curvature.

The WGS-84 radii we need:

| symbol | meaning | WGS-84 value |
|--------|---------|--------------|
| a	| semi-major axis (equatorial)	| 6 378.137 km |
| b	| semi-minor axis (polar)	| 6 356.752 km |
| e² |	first eccentricity² = (a² − b²)/a² |	0.00669438 |

References:[https://en.wikipedia.org/wiki/World_Geodetic_System](https://en.wikipedia.org/wiki/World_Geodetic_System)

References: [https://github.com/JSBSim-Team/jsbsim/issues/184](https://github.com/JSBSim-Team/jsbsim/issues/184)

### Deriving the corrected radius r(φ)


1. **Radius of curvature in the prime vertical**

``` N(φ) = a / √(1 − e² sin² φ) ```

2. **Distance from the spin axis**

``` r(φ) = N(φ) · cos φ ```

```
// RadiusOfCurvaturePrimeVertical returns N(φ):
// N(φ) = a / sqrt(1 - e² * sin²(φ))
func RadiusOfCurvaturePrimeVertical(a, e, phi float64) float64 {
	return a / math.Sqrt(1-math.Pow(e, 2)*math.Pow(math.Sin(phi), 2))
}

// CorrectedRadius returns r(φ) = N(φ) * cos(φ)
func CorrectedRadius(a, e, phi float64) float64 {
	N := RadiusOfCurvaturePrimeVertical(a, e, phi)
	return N * math.Cos(phi)
}
```

### Putting those together accepting a latitude (a and b are constants in main.go)

```
// Horizontal distance from Earth’s spin axis at latitude φ (degrees)
func radiusAtLatitude(phiDeg float64) float64 {
	phi := phiDeg * math.Pi / 180
	N := a / math.Sqrt(1-e2*math.Sin(phi)*math.Sin(phi))
	return N * math.Cos(phi) // km
}
```

### And then, we just need to translate radius to linear speed

We need to obtain the total path side real in a whole day: C(φ) and the average linear speed v(φ).
Here, C(φ) is circular:
```
C(φ)=2πr(φ)
```
So, the speed is (where T is 24h):
```
v(φ)= (2 * πr(φ)) / T
```

In go, I implemented as:
```
	r := radiusAtLatitude(loc.Lat)
	return math.Abs(2 * math.Pi * r / dayHours)
```
