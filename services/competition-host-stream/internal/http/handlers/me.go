package handlers

// me.go - placeholder for /me endpoint handlers
// GraphQL Schema Reference:
/*
  "Contains information relating to the current user."
  type Me {
    "Get info about matchmaking queues user is part of"
    matchmaking(filter: OwnMatchmakingFilterInput): MatchmakingSummary @cost(weight: "10")
    "Get tournaments which the current user is part of."
    ownTournaments(filter: OwnTournamentsInput): [Tournament!]! @cost(weight: "10") @deprecated(reason: "Use GetOwnTournamentParticipations instead.")
    "Get get tournament participations for current tournaments."
    ownTournamentParticipations(filter: OwnTournamentsInput): [OwnTournamentParticipation!]! @cost(weight: "10") @deprecated(reason: "Use 'tournaments' instead.")
    "Get get tournament participations for current tournaments."
    tournaments(filter: OwnTournamentsInput "Returns the first _n_ elements from the list." first: Int "Returns the elements in the list that come after the specified cursor." after: String "Returns the last _n_ elements from the list." last: Int "Returns the elements in the list that come before the specified cursor." before: String): TournamentsConnection @listSize(assumedSize: 30, slicingArguments: [ "first", "last" ], slicingArgumentDefaultValue: 10, sizedFields: [ "edges", "nodes" ], requireOneSlicingArgument: false) @cost(weight: "10")
    "Get user's own leaderboard participations."
    leaderboards("Returns the first _n_ elements from the list." first: Int "Returns the elements in the list that come after the specified cursor." after: String "Returns the last _n_ elements from the list." last: Int "Returns the elements in the list that come before the specified cursor." before: String): LeaderboardsConnection @listSize(assumedSize: 30, slicingArguments: [ "first", "last" ], slicingArgumentDefaultValue: 10, sizedFields: [ "edges", "nodes" ], requireOneSlicingArgument: false) @cost(weight: "10")
    "Get Ladders which the current user is part of."
    ownLadders(filter: OwnLaddersInput): [Ladder!]! @cost(weight: "10") @deprecated(reason: "Ladders are replaced by `leaderboards` throughout the API.")
    "Get participation in cups which the current user is part of."
    ownCupParticipation: [OwnCupParticipation!]! @cost(weight: "10") @deprecated(reason: "Use '\/me\/cups' or '\/me\/currentCups' instead.")
    "Get cups which the current user is part of."
    currentCups: [OwnCupParticipation!]! @cost(weight: "10")
    "Search among this user's cups."
    cups(filters: SearchOwnCupsFiltersInput! "Returns the first _n_ elements from the list." first: Int "Returns the elements in the list that come after the specified cursor." after: String): CupsConnection @listSize(assumedSize: 10, slicingArguments: [ "first" ], slicingArgumentDefaultValue: 5, sizedFields: [ "edges", "nodes" ], requireOneSlicingArgument: false) @cost(weight: "10")
    "The profile of the current user."
    user: UserProfile!
  }
*/
